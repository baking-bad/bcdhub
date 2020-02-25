package elastic

import (
	"errors"
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// GetOperationByID -
func (e *Elastic) GetOperationByID(id string) (op models.Operation, err error) {
	resp, err := e.GetByID(DocOperations, id)
	if err != nil {
		return
	}
	if !resp.Get("found").Bool() {
		return op, fmt.Errorf("Unknown operation with ID %s", id)
	}
	op.ParseElasticJSON(resp)
	return
}

// GetOperationByHash -
func (e *Elastic) GetOperationByHash(hash string) (ops []models.Operation, err error) {
	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("hash", hash),
			),
		),
	).Add(qItem{
		"sort": qItem{
			"_script": qItem{
				"type": "number",
				"script": qItem{
					"lang":   "painless",
					"source": "doc['level'].value * 10 + (doc['internal'].value ? 0 : 1)",
				},
				"order": "desc",
			},
		},
	}).All()
	resp, err := e.query(DocOperations, query)
	if err != nil {
		return
	}
	if resp.Get("hits.total.value").Int() < 1 {
		return nil, fmt.Errorf("Unknown operation with hash %s", hash)
	}
	count := resp.Get("hits.hits.#").Int()
	ops = make([]models.Operation, count)
	for i, item := range resp.Get("hits.hits").Array() {
		var o models.Operation
		o.ParseElasticJSON(item)
		ops[i] = o
	}

	return ops, nil
}

func (e *Elastic) getContractOPG(address, network string, lastID, size uint64) ([]string, error) {
	if size == 0 {
		size = 10
	}

	var lastIDFilter string
	if lastID != 0 {
		lastIDFilter = fmt.Sprintf(" AND indexed_time < %d", lastID)
	}

	sqlString := fmt.Sprintf(`SELECT hash, level
		FROM operation 
		WHERE (source = '%s' OR destination = '%s') AND network = '%s' %s 
		GROUP BY hash, level 
		ORDER BY level DESC 
		LIMIT %d`, address, address, network, lastIDFilter, size)

	res, err := e.executeSQL(sqlString)
	if err != nil {
		return nil, err
	}

	hash := make([]string, 0)
	for _, item := range res.Get("rows").Array() {
		hash = append(hash, item.Get("0").String())
	}

	return hash, nil
}

// GetContractOperations -
func (e *Elastic) GetContractOperations(network, address string, lastID, size uint64) (po PageableOperations, err error) {
	opg, err := e.getContractOPG(address, network, lastID, size)
	if err != nil {
		return
	}
	s := make([]qItem, len(opg))
	for i := range opg {
		s[i] = matchPhrase("hash", opg[i])
	}

	b := boolQ(
		should(s...),
		must(
			matchPhrase("network", network),
		),
	)
	b.Get("bool").Append("minimum_should_match", 1)
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
						"source": "doc['level'].value * 10 + (doc['internal'].value ? 0 : 1)",
					},
					"order": "desc",
				},
			},
		}).All()

	res, err := e.query(DocOperations, query)
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

// GetLastStorage -
func (e *Elastic) GetLastStorage(network, address string) (gjson.Result, error) {
	query := newQuery().
		Query(
			boolQ(
				must(
					matchPhrase("network", network),
					matchPhrase("destination", address),
				),
				notMust(
					term("deffated_storage", ""),
				),
			),
		).
		Add(qItem{
			"sort": qItem{
				"_script": qItem{
					"type": "number",
					"script": qItem{
						"lang":   "painless",
						"source": "doc['level'].value * 10 + (doc['internal'].value ? 0 : 1)",
					},
					"order": "desc",
				},
			},
		}).
		One()

	res, err := e.query(DocOperations, query)
	if err != nil {
		return gjson.Result{}, err
	}

	if res.Get("hits.total.value").Int() < 1 {
		return gjson.Result{}, nil
	}
	return res.Get("hits.hits.0"), nil
}

// GetPreviousOperation -
func (e *Elastic) GetPreviousOperation(address, network string, level int64) (op models.Operation, err error) {
	query := newQuery().
		Query(
			boolQ(
				must(
					matchPhrase("destination", address),
					matchPhrase("network", network),
					rangeQ("level", qItem{"lt": level}),
				),
				notMust(
					term("deffated_storage", ""),
				),
			),
		).
		Add(qItem{
			"sort": qItem{
				"_script": qItem{
					"type": "number",
					"script": qItem{
						"lang":   "painless",
						"source": "doc['level'].value * 10 + (doc['internal'].value ? 0 : 1)",
					},
					"order": "desc",
				},
			},
		}).One()

	res, err := e.query(DocOperations, query)
	if err != nil {
		return
	}

	if res.Get("hits.total.value").Int() < 1 {
		return op, errors.New("Operation not found")
	}
	op.ParseElasticJSON(res.Get("hits.hits.0"))
	return
}
