package elastic

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

type opgForContract struct {
	hash    string
	counter int64
}

func (e *Elastic) getContractOPG(address, network string, size uint64, filters map[string]interface{}) ([]opgForContract, error) {
	if size == 0 || size > maxQuerySize {
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

	var response sqlResponse
	if err := e.executeSQL(sqlString, &response); err != nil {
		return nil, err
	}

	resp := make([]opgForContract, 0)
	for i := range response.Rows {
		resp = append(resp, opgForContract{
			hash:    response.Rows[i][0].(string),
			counter: int64(response.Rows[i][1].(float64)),
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
				return "", errors.Errorf("Unknown operation filter: %s %v", k, v)
			}
		}
	}
	return
}

type getOperationsForContractResponse struct {
	Hist HitsArray `json:"hits"`
	Agg  struct {
		LastID floatValue `json:"last_id"`
	} `json:"aggregations"`
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
			aggs(aggItem{"last_id", min("indexed_time")}),
		).
		Add(qItem{
			"sort": qItem{
				"_script": qItem{
					"type": "number",
					"script": qItem{
						"lang":   "painless",
						"source": "doc['level'].value * 10000000000L + (doc['counter'].value) * 1000L + (doc['internal'].value ? (998L - doc['nonce'].value) : 999L)",
					},
					"order": "desc",
				},
			},
		}).All()

	var response getOperationsForContractResponse
	if err = e.query([]string{DocOperations}, query, &response); err != nil {
		return
	}

	ops := make([]models.Operation, len(response.Hist.Hits))
	for i := range response.Hist.Hits {
		if err = json.Unmarshal(response.Hist.Hits[i].Source, &ops[i]); err != nil {
			return
		}
		ops[i].ID = response.Hist.Hits[i].ID
	}

	po.Operations = ops
	po.LastID = fmt.Sprintf("%.0f", response.Agg.LastID.Value)
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

	var response SearchResponse
	if err = e.query([]string{DocOperations}, query, &response); err != nil {
		return
	}

	if response.Hits.Total.Value == 0 {
		return op, NewRecordNotFoundError(DocOperations, "", query)
	}
	err = json.Unmarshal(response.Hits.Hits[0].Source, &op)
	op.ID = response.Hits.Hits[0].ID
	return
}

type affected struct {
	addresses map[string]struct{}
}

// GetQueue -
func (a *affected) GetQueue() string {
	return ""
}

// Marshal -
func (a *affected) Marshal() ([]byte, error) {
	return nil, nil
}

// GetID -
func (a *affected) GetID() string {
	return ""
}

// GetIndex -
func (a *affected) GetIndex() string {
	return DocOperations
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

	affectedMap := affected{
		addresses: make(map[string]struct{}),
	}
	if err := e.getAllByQuery(query, &affectedMap); err != nil {
		return nil, err
	}

	addresses := make([]string, 0)
	for k := range affectedMap.addresses {
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
						"source": "doc['level'].value * 10000000000L + (doc['counter'].value) * 1000L + (doc['internal'].value ? (998L - doc['nonce'].value) : 999L)",
					},
					"order": "desc",
				},
			},
		})
	}

	scrollSize := size
	if defaultScrollSize < scrollSize || scrollSize == 0 {
		scrollSize = defaultScrollSize
	}

	ctx := newScrollContext(e, query, size, scrollSize)
	err := ctx.get(&operations)
	return operations, err
}

// OperationsStats -
type OperationsStats struct {
	Count      int64
	LastAction time.Time
}

type getOperationsStatsResponse struct {
	Aggs struct {
		OPG struct {
			Value int64 `json:"value"`
		} `json:"opg"`
		LastAction struct {
			Value time.Time `json:"value_as_string"`
		} `json:"last_action"`
	} `json:"aggregations"`
}

// GetOperationsStats -
func (e Elastic) GetOperationsStats(network, address string) (stats OperationsStats, err error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				boolQ(
					should(
						matchPhrase("source", address),
						matchPhrase("destination", address),
					),
					minimumShouldMatch(1),
				),
			),
		),
	).Add(
		aggs(
			aggItem{
				"opg", count("hash.keyword"),
			},
			aggItem{
				"last_action", max("timestamp"),
			},
		),
	).Zero()

	var response getOperationsStatsResponse
	if err = e.query([]string{DocOperations}, query, &response); err != nil {
		return
	}

	stats.Count = response.Aggs.OPG.Value
	stats.LastAction = response.Aggs.LastAction.Value
	return
}
