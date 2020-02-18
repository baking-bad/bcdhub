package elastic

import (
	"errors"
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

func parseOperation(resp gjson.Result) models.Operation {
	op := models.Operation{
		ID: resp.Get("_id").String(),

		Protocol:  resp.Get("_source.protocol").String(),
		Hash:      resp.Get("_source.hash").String(),
		Internal:  resp.Get("_source.internal").Bool(),
		Network:   resp.Get("_source.network").String(),
		Timestamp: resp.Get("_source.timestamp").Time().UTC(),

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

// GetOperationByID -
func (e *Elastic) GetOperationByID(id string) (op models.Operation, err error) {
	resp, err := e.GetByID(DocOperations, id)
	if err != nil {
		return
	}
	if !resp.Get("found").Bool() {
		return op, fmt.Errorf("Unknown contract with ID %s", id)
	}
	op = parseOperation(*resp)
	return
}

// GetContractOperations -
func (e *Elastic) GetContractOperations(network, address string, offset, size int64) ([]models.Operation, error) {
	if size == 0 {
		size = 10
	}

	b := boolQ(
		should(
			matchPhrase("source", address),
			matchPhrase("destination", address),
		),
		must(
			matchPhrase("network", network),
		),
	)
	b.Get("bool").Append("minimum_should_match", 1)
	query := newQuery().
		Query(b).
		Size(size).
		From(offset).
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
		})

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
func (e *Elastic) GetPreviousOperation(address, network string, level int64) (models.Operation, error) {
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
		return models.Operation{}, err
	}

	if res.Get("hits.total.value").Int() < 1 {
		return models.Operation{}, errors.New("Operation not found")
	}
	return parseOperation(res.Get("hits.hits.0")), nil
}
