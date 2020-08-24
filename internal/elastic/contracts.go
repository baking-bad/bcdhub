package elastic

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

func filtersToQuery(by map[string]interface{}) base {
	matches := make([]qItem, 0)
	for k, v := range by {
		matches = append(matches, matchPhrase(k, v))
	}
	return newQuery().Query(
		boolQ(
			must(matches...),
		),
	)
}

func (e *Elastic) getContract(q map[string]interface{}) (c models.Contract, err error) {
	res, err := e.query([]string{DocContracts}, q)
	if err != nil {
		return
	}
	if res.Get("hits.total.value").Int() < 1 {
		return c, fmt.Errorf("%s: %v", RecordNotFound, q)
	}
	hit := res.Get("hits.hits.0")
	c.ParseElasticJSON(hit)
	return
}

func (e *Elastic) getContracts(query base) ([]models.Contract, error) {
	contracts := make([]models.Contract, 0)
	if err := e.getAllByQuery(query, &contracts); err != nil {
		return nil, err
	}

	return contracts, nil
}

// GetContract -
func (e *Elastic) GetContract(by map[string]interface{}) (models.Contract, error) {
	query := filtersToQuery(by).One()
	return e.getContract(query)
}

// GetContracts -
func (e *Elastic) GetContracts(by map[string]interface{}) ([]models.Contract, error) {
	query := filtersToQuery(by)
	return e.getContracts(query)
}

// GetContractRandom -
func (e *Elastic) GetContractRandom() (models.Contract, error) {
	random := qItem{
		"function_score": qItem{
			"functions": []qItem{
				{
					"random_score": qItem{
						"seed": time.Now().UnixNano(),
					},
				},
			},
		},
	}

	txRange := rangeQ("tx_count", qItem{
		"gte": 2,
	})
	b := boolQ(must(txRange, random))
	query := newQuery().Query(b).One()
	return e.getContract(query)
}

// GetContractWithdrawn -
func (e *Elastic) GetContractWithdrawn(address, network string) (int64, error) {
	b := boolQ(
		filter(
			matchQ("network", network),
			matchQ("source", address),
		),
	)
	query := newQuery().Query(b).Add(
		qItem{
			"aggs": qItem{
				"total_withdrawn": sum("amount"),
			},
		},
	).Zero()
	res, err := e.query([]string{DocOperations}, query)
	if err != nil {
		return 0, err
	}

	return res.Get("aggregations.total_withdrawn.value").Int(), nil
}

// IsFAContract -
func (e *Elastic) IsFAContract(network, address string) (bool, error) {
	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("network", network),
				matchPhrase("address", address),
			),
			filter(
				qItem{
					"terms": qItem{
						"tags": []string{"fa12", "fa1"},
					},
				},
			),
		),
	)
	resp, err := e.query([]string{DocContracts}, query, "address")
	if err != nil {
		return false, err
	}
	return resp.Get("hits.total.value").Int() == 1, nil
}

// UpdateContractMigrationsCount -
func (e *Elastic) UpdateContractMigrationsCount(address, network string) error {
	contract, err := e.GetContract(map[string]interface{}{
		"address": address,
		"network": network,
	})
	if err != nil {
		return err
	}
	contract.MigrationsCount++

	_, err = e.UpdateDoc(DocContracts, contract.ID, contract)
	return err
}

// GetContractAddressesByNetworkAndLevel -
func (e *Elastic) GetContractAddressesByNetworkAndLevel(network string, maxLevel int64) (gjson.Result, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				rangeQ("level", qItem{
					"gt": maxLevel,
				}),
			),
		),
	).All()
	resp, err := e.query([]string{DocContracts}, query, "address")
	if err != nil {
		return resp, err
	}
	return resp.Get("hits.hits"), nil
}

// NeedParseOperation -
func (e *Elastic) NeedParseOperation(network, source, destination string) (bool, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				boolQ(
					should(
						matchPhrase("address", source),
						matchPhrase("address", destination),
					),
					minimumShouldMatch(1),
				),
			),
		),
	).One()
	resp, err := e.query([]string{DocContracts}, query, "address")
	if err != nil {
		return false, err
	}
	return resp.Get("hits.total.value").Int() == 1, nil
}

type contractIDs struct {
	IDs []string
}

// GetQueue -
func (ids *contractIDs) GetQueue() string {
	return ""
}

// GetID -
func (ids *contractIDs) GetID() string {
	return ""
}

// GetIndex -
func (ids *contractIDs) GetIndex() string {
	return DocContracts
}

// ParseElasticJSON -
func (ids *contractIDs) ParseElasticJSON(hit gjson.Result) {
	ids.IDs = append(ids.IDs, hit.Get("_id").String())
}

// GetContractsIDByAddress -
func (e *Elastic) GetContractsIDByAddress(addresses []string, network string) ([]string, error) {
	shouldItems := make([]qItem, len(addresses))
	for i := range addresses {
		shouldItems[i] = matchPhrase("address", addresses[i])
	}

	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
			should(shouldItems...),
			minimumShouldMatch(1),
		),
	)

	ids := contractIDs{
		IDs: make([]string, 0),
	}
	err := e.getAllByQuery(query, &ids)
	return ids.IDs, err
}

// RecalcContractStats -
func (e *Elastic) RecalcContractStats(network, address string) (stats ContractStats, err error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
			should(
				matchPhrase("source", address),
				matchPhrase("destination", address),
			),
			minimumShouldMatch(1),
		),
	).Add(
		qItem{
			"aggs": qItem{
				"tx_count":    count("indexed_time"),
				"last_action": max("timestamp"),
				"balance": qItem{
					"scripted_metric": qItem{
						"init_script":    "state.operations = []",
						"map_script":     "if (doc['status.keyword'].value == 'applied' && doc['amount'].size() != 0) {state.operations.add(doc['destination.keyword'].value == params.address ? doc['amount'].value : -1L * doc['amount'].value)}",
						"combine_script": "double balance = 0; for (amount in state.operations) { balance += amount } return balance",
						"reduce_script":  "double balance = 0; for (a in states) { balance += a } return balance",
						"params": qItem{
							"address": address,
						},
					},
				},
				"total_withdrawn": qItem{
					"scripted_metric": qItem{
						"init_script":    "state.operations = []",
						"map_script":     "if (doc['status.keyword'].value == 'applied' && doc['amount'].size() != 0 && doc['source.keyword'].value == params.address) {state.operations.add(doc['amount'].value)}",
						"combine_script": "double balance = 0; for (amount in state.operations) { balance += amount } return balance",
						"reduce_script":  "double balance = 0; for (a in states) { balance += a } return balance",
						"params": qItem{
							"address": address,
						},
					},
				},
			},
		},
	).Zero()
	response, err := e.query([]string{DocOperations}, query)
	if err != nil {
		return
	}

	stats.ParseElasticJSON(response.Get("aggregations"))

	return
}

// GetContractMigrationStats -
func (e *Elastic) GetContractMigrationStats(network, address string) (stats ContractMigrationsStats, err error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
			should(
				matchPhrase("source", address),
				matchPhrase("destination", address),
			),
			minimumShouldMatch(1),
		),
	).Add(
		qItem{
			"aggs": qItem{
				"migrations_count": count("indexed_time"),
			},
		},
	).Zero()
	response, err := e.query([]string{DocMigrations}, query)
	if err != nil {
		return
	}

	stats.ParseElasticJSON(response.Get("aggregations"))

	return
}

// GetDAppStats -
func (e *Elastic) GetDAppStats(network string, addresses []string, period string) (stats DAppStats, err error) {
	addressMatches := make([]qItem, len(addresses))
	for i := range addresses {
		addressMatches[i] = matchPhrase("destination", addresses[i])
	}

	matches := []qItem{
		matchQ("network", network),
		boolQ(
			should(addressMatches...),
			minimumShouldMatch(1),
		),
		matchQ("status", "applied"),
	}
	r, err := periodToRange(period)
	if err != nil {
		return
	}
	if r != nil {
		matches = append(matches, r)
	}

	query := newQuery().Query(
		boolQ(
			filter(matches...),
		),
	).Add(
		qItem{
			"aggs": qItem{
				"users": qItem{
					"cardinality": qItem{
						"field": "source.keyword",
					},
				},
				"tx":     count("indexed_time"),
				"volume": sum("amount"),
			},
		},
	).Zero()

	response, err := e.query([]string{DocOperations}, query)
	if err != nil {
		return
	}

	stats.ParseElasticJSON(response.Get("aggregations"))
	return
}

func periodToRange(period string) (qItem, error) {
	var str string
	switch period {
	case "year":
		str = "now-1y/d"
	case "month":
		str = "now-1M/d"
	case "week":
		str = "now-1w/d"
	case "day":
		str = "now-1d/d"
	case "all":
		return nil, nil
	default:
		return nil, fmt.Errorf("Unknown period value: %s", period)
	}
	return qItem{
		"range": qItem{
			"timestamp": qItem{
				"gte": str,
			},
		},
	}, nil
}

// GetContractTransfers -
func (e *Elastic) GetContractTransfers(network string, address string, size, offset int64) (TransfersResponse, error) {
	matches := []qItem{
		matchQ("network", network),
		matchPhrase("contract", address),
	}

	if size == 0 {
		size = defaultSize
	}

	query := newQuery().Query(
		boolQ(
			filter(matches...),
		),
	).Sort("indexed_time", "desc").Size(size).From(offset)

	response, err := e.query([]string{DocTransfers}, query)
	if err != nil {
		return TransfersResponse{}, err
	}

	transfers := make([]models.Transfer, 0)
	hits := response.Get("hits.hits").Array()
	for _, hit := range hits {
		var transfer models.Transfer
		transfer.ParseElasticJSON(hit)
		transfers = append(transfers, transfer)
	}
	return TransfersResponse{
		Transfers: transfers,
		Total:     response.Get("hits.total.value").Int(),
	}, nil
}
