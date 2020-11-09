package elastic

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
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

func (e *Elastic) getContract(q base) (c models.Contract, err error) {
	var response SearchResponse
	if err = e.query([]string{DocContracts}, q, &response); err != nil {
		return
	}
	if response.Hits.Total.Value == 0 {
		return c, NewRecordNotFoundError(DocContracts, "", q)
	}
	err = json.Unmarshal(response.Hits.Hits[0].Source, &c)
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
	var response SearchResponse
	if err := e.query([]string{DocContracts}, query, &response, "address"); err != nil {
		return false, err
	}
	return response.Hits.Total.Value == 1, nil
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
	return e.UpdateDoc(&contract)
}

// GetContractAddressesByNetworkAndLevel -
func (e *Elastic) GetContractAddressesByNetworkAndLevel(network string, maxLevel int64) ([]string, error) {
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

	var response SearchResponse
	if err := e.query([]string{DocContracts}, query, "address"); err != nil {
		return nil, err
	}

	addresses := make([]string, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		var c models.Contract
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &c); err != nil {
			return nil, err
		}
		addresses[i] = c.Address
	}

	return addresses, nil
}

type contractIDs struct {
	IDs []string
}

// GetQueue -
func (ids *contractIDs) GetQueue() string {
	return ""
}

// Marshal -
func (ids *contractIDs) Marshal() ([]byte, error) {
	return nil, nil
}

// GetID -
func (ids *contractIDs) GetID() string {
	return ""
}

// GetIndex -
func (ids *contractIDs) GetIndex() string {
	return DocContracts
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

type recalcContractStatsResponse struct {
	Aggs struct {
		TxCount struct {
			Value int64 `json:"value"`
		} `json:"tx_count"`
		Balance struct {
			Value int64 `json:"value"`
		} `json:"balance"`
		LastAction struct {
			Value int64 `json:"value"`
		} `json:"last_action"`
		TotalWithdrawn struct {
			Value int64 `json:"value"`
		} `json:"total_withdrawn"`
	} `json:"aggregations"`
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
	var response recalcContractStatsResponse
	if err = e.query([]string{DocOperations}, query, &response); err != nil {
		return
	}

	stats.LastAction = time.Unix(0, response.Aggs.LastAction.Value*1000000).UTC()
	stats.Balance = response.Aggs.Balance.Value
	stats.TotalWithdrawn = response.Aggs.TotalWithdrawn.Value
	stats.TxCount = response.Aggs.TxCount.Value
	return
}

type getContractMigrationStatsResponse struct {
	Agg struct {
		MigrationsCount struct {
			Value int64 `json:"value"`
		} `json:"migrations_count"`
	} `json:"aggregations"`
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
		aggs(
			aggItem{
				"migrations_count", count("indexed_time"),
			},
		),
	).Zero()

	var response getContractMigrationStatsResponse
	if err = e.query([]string{DocMigrations}, query, &response); err != nil {
		return
	}

	stats.MigrationsCount = response.Agg.MigrationsCount.Value
	return
}

type getDAppStatsResponse struct {
	Aggs struct {
		Users struct {
			Value float64 `json:"value"`
		} `json:"users"`
		Calls struct {
			Value float64 `json:"value"`
		} `json:"calls"`
		Volume struct {
			Value float64 `json:"value"`
		} `json:"volume"`
	} `json:"aggregations"`
}

// GetDAppStats -
func (e *Elastic) GetDAppStats(network string, addresses []string, period string) (stats DAppStats, err error) {
	addressMatches := make([]qItem, len(addresses))
	for i := range addresses {
		addressMatches[i] = matchPhrase("destination", addresses[i])
	}

	matches := []qItem{
		matchQ("network", network),
		exists("entrypoint"),
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
		aggs(
			aggItem{"users", cardinality("source.keyword")},
			aggItem{"calls", count("indexed_time")},
			aggItem{"volume", sum("amount")},
		),
	).Zero()

	var response getDAppStatsResponse
	if err = e.query([]string{DocOperations}, query, &response); err != nil {
		return
	}

	stats.Calls = int64(response.Aggs.Calls.Value)
	stats.Users = int64(response.Aggs.Users.Value)
	stats.Volume = int64(response.Aggs.Volume.Value)
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
		return nil, errors.Errorf("Unknown period value: %s", period)
	}
	return qItem{
		"range": qItem{
			"timestamp": qItem{
				"gte": str,
			},
		},
	}, nil
}

// GetContractsByAddresses -
func (e *Elastic) GetContractsByAddresses(addresses []Address) ([]models.Contract, error) {
	items := make([]qItem, len(addresses))
	for i := range addresses {
		items[i] = boolQ(
			filter(
				matchPhrase("address", addresses[i].Address),
				matchQ("network", addresses[i].Network),
			),
		)
	}

	query := newQuery().Query(
		boolQ(
			should(items...),
			minimumShouldMatch(1),
		),
	)
	contracts := make([]models.Contract, 0)
	err := e.getAllByQuery(query, &contracts)
	return contracts, err
}
