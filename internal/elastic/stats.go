package elastic

// GetNetworkCountStats -
func (e *Elastic) GetNetworkCountStats(network string) (map[string]int64, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
			should(
				exists("entrypoint"),
				exists("fingerprint"),
			),
			minimumShouldMatch(1),
		),
	).Add(
		aggs(
			aggItem{
				"body",
				termsAgg("_index", maxQuerySize),
			},
		),
	).Zero()

	return e.getCountAgg([]string{DocContracts, DocOperations}, query)
}

// GetCallsCountByNetwork -
func (e *Elastic) GetCallsCountByNetwork() (map[string]int64, error) {
	query := newQuery().Query(exists("entrypoint")).Add(
		aggs(
			aggItem{
				"body", qItem{
					"terms": qItem{
						"field": "network.keyword",
					},
				},
			},
		),
	).Zero()

	return e.getCountAgg([]string{DocOperations}, query)
}

type getContractStatsByNetworkStats struct {
	Agg struct {
		Network struct {
			Buckets []struct {
				Bucket
				Same           intValue `json:"same"`
				Balance        intValue `json:"balance"`
				TotalWithdrawn intValue `json:"total_withdrawn"`
			} `json:"buckets"`
		} `json:"network"`
	} `json:"aggregations"`
}

// GetContractStatsByNetwork -
func (e *Elastic) GetContractStatsByNetwork() (map[string]ContractCountStats, error) {
	query := newQuery().Add(
		aggs(
			aggItem{
				"network", qItem{
					"terms": qItem{
						"field": "network.keyword",
					},
					"aggs": qItem{
						"same": qItem{
							"cardinality": qItem{
								"script": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value",
							},
						},
						"balance":         sum("balance"),
						"total_withdrawn": sum("total_withdrawn"),
					},
				},
			},
		),
	).Zero()

	var response getContractStatsByNetworkStats
	if err := e.query([]string{DocContracts}, query, &response); err != nil {
		return nil, err
	}

	counts := make(map[string]ContractCountStats)
	for _, item := range response.Agg.Network.Buckets {
		counts[item.Key] = ContractCountStats{
			Total:          item.DocCount,
			SameCount:      item.Same.Value,
			Balance:        item.Balance.Value,
			TotalWithdrawn: item.TotalWithdrawn.Value,
		}
	}
	return counts, nil
}

// GetFACountByNetwork -
func (e *Elastic) GetFACountByNetwork() (map[string]int64, error) {
	query := newQuery().Query(
		in("tags", []string{
			"fa1",
			"fa12",
		}),
	).Add(
		aggs(
			aggItem{
				"body", qItem{
					"terms": qItem{
						"field": "network.keyword",
					},
				},
			},
		),
	).Zero()

	return e.getCountAgg([]string{DocContracts}, query)
}

// GetLanguagesForNetwork -
func (e *Elastic) GetLanguagesForNetwork(network string) (map[string]int64, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
		),
	).Add(
		aggs(
			aggItem{
				"body", qItem{
					"terms": qItem{
						"field": "language.keyword",
					},
				},
			},
		),
	).Zero()

	return e.getCountAgg([]string{DocContracts}, query)
}
