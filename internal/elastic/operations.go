package elastic

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
)

type opgForContract struct {
	hash    string
	counter int64
}

func (e *Elastic) getContractOPG(address, network string, size uint64, filters map[string]interface{}) ([]opgForContract, error) {
	if size == 0 {
		size = defaultSize
	}

	filtersString, err := prepareOperationFilters(filters)
	if err != nil {
		return nil, err
	}

	sqlString := fmt.Sprintf(`SELECT hash, counter
		FROM operation 
		WHERE (source = '%s' OR destination = '%s') AND network = '%s' %s 
		GROUP BY hash, counter, level
		ORDER BY level DESC
		LIMIT %d`, address, address, network, filtersString, size)

	res, err := e.executeSQL(sqlString)
	if err != nil {
		return nil, err
	}

	resp := make([]opgForContract, 0)
	for _, item := range res.Get("rows").Array() {
		resp = append(resp, opgForContract{
			hash:    item.Get("0").String(),
			counter: item.Get("1").Int(),
		})
	}

	return resp, nil
}

func prepareOperationFilters(filters map[string]interface{}) (s string, err error) {
	for k, v := range filters {
		if v != "" {
			s += " AND "
			switch k {
			case "from":
				s += fmt.Sprintf("timestamp >= %d", v)
			case "to":
				s += fmt.Sprintf("timestamp <= %d", v)
			case "entrypoints":
				s += fmt.Sprintf("entrypoint IN (%s)", v)
			case "last_id":
				s += fmt.Sprintf("indexed_time < %s", v)
			case "status":
				s += fmt.Sprintf("status IN (%s)", v)
			default:
				return "", fmt.Errorf("Unknown operation filter: %s %v", k, v)
			}
		}
	}
	return
}

// GetOperationsForContract -
func (e *Elastic) GetOperationsForContract(network, address string, size uint64, filters map[string]interface{}) (po PageableOperations, err error) {
	opg, err := e.getContractOPG(address, network, size, filters)
	if err != nil {
		return
	}

	s := make([]qItem, len(opg))
	for i := range opg {
		s[i] = boolQ(filter(
			matchQ("hash", opg[i].hash),
			term("counter", opg[i].counter),
		))
	}
	b := boolQ(
		should(s...),
		filter(
			matchQ("network", network),
		),
		minimumShouldMatch(1),
	)
	query := newQuery().
		Query(b).
		Add(
			aggs("last_id", min("indexed_time")),
		).
		Add(qItem{
			"sort": qItem{
				"_script": qItem{
					"type": "number",
					"script": qItem{
						"lang":   "painless",
						"inline": "doc['level'].value * 10000000000L + (doc['counter'].value) * 1000L + (doc['internal'].value ? (998L - doc['nonce'].value) : 999L)",
					},
					"order": "desc",
				},
			},
		}).All()

	res, err := e.query([]string{DocOperations}, query)
	if err != nil {
		return
	}

	count := res.Get("hits.hits.#").Int()
	ops := make([]models.Operation, count)
	for i, item := range res.Get("hits.hits").Array() {
		var o models.Operation
		o.ParseElasticJSON(item)
		ops[i] = o
	}

	po.Operations = ops
	po.LastID = res.Get("aggregations.last_id.value").String()

	return
}

// GetLastOperation -
func (e *Elastic) GetLastOperation(address, network string, indexedTime int64) (op models.Operation, err error) {
	query := newQuery().
		Query(
			boolQ(
				must(
					matchPhrase("destination", address),
					matchPhrase("network", network),
				),
				filter(
					rangeQ("indexed_time", qItem{"lt": indexedTime}),
					term("status", "applied"),
				),
				notMust(
					term("deffated_storage", ""),
				),
			),
		).Sort("indexed_time", "desc").One()

	res, err := e.query([]string{DocOperations}, query)
	if err != nil {
		return
	}

	if res.Get("hits.total.value").Int() < 1 {
		return op, fmt.Errorf("%s %s in %s on %d", RecordNotFound, address, network, indexedTime)
	}
	op.ParseElasticJSON(res.Get("hits.hits.0"))
	return
}

// GetAllLevelsForNetwork -
func (e *Elastic) GetAllLevelsForNetwork(network string) (map[int64]struct{}, error) {
	levels := make(map[int64]struct{})

	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
		),
	).Sort("level", "asc")
	result, err := e.createScroll(DocOperations, 1000, query)
	if err != nil {
		return nil, err
	}
	for {
		scrollID := result.Get("_scroll_id").String()
		hits := result.Get("hits.hits")
		if hits.Get("#").Int() < 1 {
			break
		}

		for _, item := range hits.Array() {
			level := item.Get("_source.level").Int()
			levels[level] = struct{}{}
		}

		result, err = e.queryScroll(scrollID)
		if err != nil {
			return nil, err
		}
	}

	return levels, nil
}

// GetAffectedContracts -
func (e *Elastic) GetAffectedContracts(network string, fromLevel, toLevel int64) ([]string, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				rangeQ("level", qItem{
					"lte": fromLevel,
					"gt":  toLevel,
				}),
			),
		),
	)

	result, err := e.createScroll(DocOperations, 1000, query)
	if err != nil {
		return nil, err
	}

	addressesMap := make(map[string]struct{})
	for {
		scrollID := result.Get("_scroll_id").String()
		hits := result.Get("hits.hits")
		if hits.Get("#").Int() < 1 {
			break
		}

		for _, item := range hits.Array() {
			source := item.Get("_source.source").String()
			destination := item.Get("_source.destination").String()
			if strings.HasPrefix(source, "KT") {
				addressesMap[source] = struct{}{}
			}
			if strings.HasPrefix(destination, "KT") {
				addressesMap[destination] = struct{}{}
			}
		}

		result, err = e.queryScroll(scrollID)
		if err != nil {
			return nil, err
		}
	}

	addresses := make([]string, 0)
	for k := range addressesMap {
		addresses = append(addresses, k)
	}

	return addresses, nil
}

// GetOperations -
func (e *Elastic) GetOperations(filters map[string]interface{}, size int64, sort bool) ([]models.Operation, error) {
	operations := make([]models.Operation, 0)

	query := filtersToQuery(filters)

	if sort {
		query.Add(qItem{
			"sort": qItem{
				"_script": qItem{
					"type": "number",
					"script": qItem{
						"lang":   "painless",
						"inline": "doc['level'].value * 10000000000L + (doc['counter'].value) * 1000L + (doc['internal'].value ? (998L - doc['nonce'].value) : 999L)",
					},
					"order": "desc",
				},
			},
		})
	}

	if size == 1 {
		query = query.One()
	} else if size == 0 {
		query = query.All()
	} else {
		query = query.Size(size).All()
	}

	result, err := e.createScroll(DocOperations, 1000, query)
	if err != nil {
		return nil, err
	}
	for {
		scrollID := result.Get("_scroll_id").String()
		hits := result.Get("hits.hits")
		if hits.Get("#").Int() < 1 {
			break
		}

		for _, item := range hits.Array() {
			var op models.Operation
			op.ParseElasticJSON(item)
			operations = append(operations, op)
		}

		result, err = e.queryScroll(scrollID)
		if err != nil {
			return nil, err
		}
	}

	return operations, nil
}
