package elastic

import (
	"fmt"
	"math"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// GetContractByAddressAndNetwork -
func (e *Elastic) GetContractByAddressAndNetwork(network, address string) (models.Contract, error) {
	return e.GetContract(map[string]interface{}{
		"address": address,
		"network": network,
	})
}

func getContractQuery(by map[string]interface{}) base {
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

func (e *Elastic) getContracts(q map[string]interface{}) ([]models.Contract, error) {
	contracts := make([]models.Contract, 0)

	result, err := e.createScroll(DocContracts, 1000, q)
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
			var c models.Contract
			c.ParseElasticJSON(item)
			contracts = append(contracts, c)
		}

		result, err = e.queryScroll(scrollID)
		if err != nil {
			return nil, err
		}
	}

	return contracts, nil
}

// GetContract -
func (e *Elastic) GetContract(by map[string]interface{}) (models.Contract, error) {
	query := getContractQuery(by).One()
	return e.getContract(query)
}

// GetContractsByIDsWithSort -
func (e *Elastic) GetContractsByIDsWithSort(ids []string, sortField, sortDirection string) ([]models.Contract, error) {
	query := newQuery().Query(
		qItem{
			"ids": qItem{
				"values": ids,
			},
		},
	).Sort(sortField, sortDirection).All()
	resp, err := e.query([]string{DocContracts}, query)
	if err != nil {
		return nil, err
	}

	if resp.Get("hits.total.value").Int() < 1 {
		return nil, fmt.Errorf("Unknown contracts with IDs %s", ids)
	}

	contracts := make([]models.Contract, 0)
	arr := resp.Get("hits.hits").Array()
	for i := range arr {
		var c models.Contract
		c.ParseElasticJSON(arr[i])
		contracts = append(contracts, c)
	}
	return contracts, nil
}

// GetContracts -
func (e *Elastic) GetContracts(by map[string]interface{}) ([]models.Contract, error) {
	query := getContractQuery(by)
	return e.getContracts(query)
}

// GetRandomContract -
func (e *Elastic) GetRandomContract() (models.Contract, error) {
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

// GetContractID -
func (e *Elastic) GetContractID(by map[string]interface{}) (string, error) {
	query := getContractQuery(by).One()
	cntr, err := e.getContract(query)
	if err != nil {
		return "", err
	}

	return cntr.ID, nil
}

// Recommendations -
func (e *Elastic) Recommendations(tags []string, language string, blackList []string, size int64) ([]models.Contract, error) {
	tagFilters := make([]qItem, len(tags))
	for i := range tags {
		tagFilters[i] = matchPhrase("tags", tags[i])
	}

	blackListFilters := make([]qItem, len(blackList))
	for i := range blackList {
		blackListFilters[i] = matchPhrase("address", blackList[i])
	}

	tagFilters = append(tagFilters, matchPhrase("language", language))
	b := boolQ(
		should(tagFilters...),
		notMust(
			blackListFilters...,
		),
		minimumShouldMatch(int(math.Min(2, float64(len(tagFilters)+1)))),
	)

	query := newQuery().Query(b).Size(size).Sort("last_action", "desc")
	resp, err := e.query([]string{DocContracts}, query)
	if err != nil {
		return nil, err
	}

	contracts := make([]models.Contract, 0)
	arr := resp.Get("hits.hits").Array()
	for i := range arr {
		var c models.Contract
		c.ParseElasticJSON(arr[i])
		contracts = append(contracts, c)
	}
	return contracts, nil
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

// IsKnownContract -
func (e *Elastic) IsKnownContract(network, address string) (bool, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				matchPhrase("address", address),
			),
		),
	).One()
	resp, err := e.query([]string{DocContracts}, query, "address")
	if err != nil {
		return false, err
	}
	return resp.Get("hits.total.value").Int() == 1, nil
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
	result, err := e.createScroll(DocContracts, 1000, query)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0)
	for {
		scrollID := result.Get("_scroll_id").String()
		hits := result.Get("hits.hits")
		if hits.Get("#").Int() < 1 {
			break
		}

		for _, item := range hits.Array() {
			ids = append(ids, item.Get("_id").String())
		}

		result, err = e.queryScroll(scrollID)
		if err != nil {
			return nil, err
		}
	}

	return ids, nil
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
						"map_script":     fmt.Sprintf("if (doc['status.keyword'].value == 'applied' && doc['amount'].size() != 0) {state.operations.add(doc['destination.keyword'].value == '%s' ? doc['amount'].value : -1L * doc['amount'].value)}", address),
						"combine_script": "double balance = 0; for (amount in state.operations) { balance += amount } return balance",
						"reduce_script":  "double balance = 0; for (a in states) { balance += a } return balance",
					},
				},
				"total_withdrawn": qItem{
					"scripted_metric": qItem{
						"init_script":    "state.operations = []",
						"map_script":     fmt.Sprintf("if (doc['status.keyword'].value == 'applied' && doc['amount'].size() != 0 && doc['source.keyword'].value == '%s') {state.operations.add(doc['amount'].value)}", address),
						"combine_script": "double balance = 0; for (amount in state.operations) { balance += amount } return balance",
						"reduce_script":  "double balance = 0; for (a in states) { balance += a } return balance",
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
