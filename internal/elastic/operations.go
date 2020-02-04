package elastic

import (
	"errors"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

func parseOperation(resp gjson.Result) models.Operation {
	op := models.Operation{
		ID: resp.Get("_id").String(),

		Protocol: resp.Get("_source.protocol").String(),
		Hash:     resp.Get("_source.hash").String(),
		Internal: resp.Get("_source.internal").Bool(),
		Network:  resp.Get("_source.network").String(),

		Level:         resp.Get("_source.level").Int(),
		Kind:          resp.Get("_source.kind").String(),
		Source:        resp.Get("_source.source").String(),
		Fee:           resp.Get("_source.fee").Int(),
		Counter:       resp.Get("_source.counter").Int(),
		GasLimit:      resp.Get("_source.gas_limit").Int(),
		StorageLimit:  resp.Get("_source.storage_limit").Int(),
		Amount:        resp.Get("_source.amount").Int(),
		Destination:   resp.Get("_source.destination").String(),
		PublicKey:     resp.Get("_source.public_key").String(),
		ManagerPubKey: resp.Get("_source.manager_pubkey").String(),
		Balance:       resp.Get("_source.balance").Int(),
		Delegate:      resp.Get("_source.delegate").String(),
		Parameters:    resp.Get("_source.parameters").String(),

		Result:         parseOperationResult(resp.Get("_source.result")),
		BalanceUpdates: make([]models.BalanceUpdate, 0),

		DeffatedStorage: resp.Get("_source.deffated_storage").String(),
	}

	for _, b := range resp.Get("_source.balance_updates").Array() {
		op.BalanceUpdates = append(op.BalanceUpdates, parseBalanceUpdate(b))
	}

	return op
}

func parseBalanceUpdate(data gjson.Result) models.BalanceUpdate {
	return models.BalanceUpdate{
		Kind:     data.Get("kind").String(),
		Contract: data.Get("contract").String(),
		Change:   data.Get("change").Int(),
		Category: data.Get("category").String(),
		Delegate: data.Get("delegate").String(),
		Cycle:    int(data.Get("cycle").Int()),
	}
}

func parseOperationResult(data gjson.Result) *models.OperationResult {
	bu := make([]models.BalanceUpdate, 0)
	for _, b := range data.Get("balance_updates").Array() {
		bu = append(bu, parseBalanceUpdate(b))
	}
	return &models.OperationResult{
		Status:              data.Get("status").String(),
		ConsumedGas:         data.Get("consumed_gas").Int(),
		StorageSize:         data.Get("storage_size").Int(),
		PaidStorageSizeDiff: data.Get("paid_storage_size_diff").Int(),
		Errors:              data.Get("errors").String(),

		BalanceUpdates: bu,
	}
}

// GetContractOperations -
func (e *Elastic) GetContractOperations(network, address string, offset, size int64) ([]models.Operation, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					map[string]interface{}{
						"match": map[string]interface{}{
							"source": address,
						},
					}, map[string]interface{}{
						"match": map[string]interface{}{
							"destination": address,
						},
					},
				},
				"must": map[string]interface{}{
					"term": map[string]interface{}{
						"network": network,
					},
				},
				"minimum_should_match": 1,
			},
		},
		"sort": map[string]interface{}{
			"_script": map[string]interface{}{
				"type": "number",
				"script": map[string]interface{}{
					"lang":   "painless",
					"source": "doc['level'].value * 10 + (doc['internal'].value ? 0 : 1)",
				},
				"order": "desc",
			},
		},
		"from": offset,
	}
	if size == 0 {
		size = 10
	}
	query["size"] = size

	res, err := e.query(DocOperations, query)
	if err != nil {
		return nil, err
	}

	ops := make([]models.Operation, 0)
	for _, item := range res.Get("hits.hits").Array() {
		ops = append(ops, parseOperation(item))
	}

	return ops, nil
}

// GetLastStorage -
func (e *Elastic) GetLastStorage(network, address string) (gjson.Result, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					map[string]interface{}{
						"match": map[string]interface{}{
							"source": address,
						},
					}, map[string]interface{}{
						"match": map[string]interface{}{
							"destination": address,
						},
					},
				},
				"must": map[string]interface{}{
					"term": map[string]interface{}{
						"network": network,
					},
				},
				"must_not": map[string]interface{}{
					"term": map[string]interface{}{
						"deffated_storage": "",
					},
				},
				"minimum_should_match": 1,
			},
		},
		"sort": map[string]interface{}{
			"_script": map[string]interface{}{
				"type": "number",
				"script": map[string]interface{}{
					"lang":   "painless",
					"source": "doc['level'].value * 10 + (doc['internal'].value ? 0 : 1)",
				},
				"order": "desc",
			},
		},
		"size": 1,
	}

	res, err := e.query(DocOperations, query)
	if err != nil {
		return gjson.Result{}, err
	}

	if res.Get("hits.total.value").Int() < 1 {
		return gjson.Result{}, nil
	}
	val := res.Get("hits.hits.0._source").Get("deffated_storage").String()
	return gjson.Parse(val), nil
}

// GetPreviousOperation -
func (e *Elastic) GetPreviousOperation(address, network string, level int64) (models.Operation, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					map[string]interface{}{
						"match": map[string]interface{}{
							"source": address,
						},
					}, map[string]interface{}{
						"match": map[string]interface{}{
							"destination": address,
						},
					},
				},
				"must": []map[string]interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"network": network,
						}},
					map[string]interface{}{
						"range": map[string]interface{}{
							"level": map[string]interface{}{
								"lt": level,
							},
						}},
				},
				"must_not": map[string]interface{}{
					"term": map[string]interface{}{
						"deffated_storage": "",
					},
				},
				"minimum_should_match": 1,
			},
		},
		"sort": map[string]interface{}{
			"_script": map[string]interface{}{
				"type": "number",
				"script": map[string]interface{}{
					"lang":   "painless",
					"source": "doc['level'].value * 10 + (doc['internal'].value ? 0 : 1)",
				},
				"order": "desc",
			},
		},
		"size": 1,
	}
	res, err := e.query(DocOperations, query)
	if err != nil {
		return models.Operation{}, err
	}

	if res.Get("hits.total.value").Int() < 1 {
		return models.Operation{}, errors.New("Operation not found")
	}
	return parseOperation(res.Get("hits.hits.0")), nil
}
