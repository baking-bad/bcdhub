package core

import (
	"github.com/baking-bad/bcdhub/internal/models"
)

// GetNetworkCountStats -
func (e *Elastic) GetNetworkCountStats(network string) (map[string]int64, error) {
	query := NewQuery().Query(
		Bool(
			Filter(
				Match("network", network),
			),
			Should(
				Exists("entrypoint"),
				Exists("fingerprint"),
			),
			MinimumShouldMatch(1),
		),
	).Add(
		Aggs(
			AggItem{
				"body",
				TermsAgg("_index", MaxQuerySize),
			},
		),
	).Zero()

	return e.GetCountAgg([]string{models.DocContracts, models.DocOperations}, query)
}

// GetCallsCountByNetwork -
func (e *Elastic) GetCallsCountByNetwork(network string) (map[string]int64, error) {
	filters := []Item{
		Exists("entrypoint"),
	}
	if network != "" {
		filters = append(filters, Match("network", network))
	}

	query := NewQuery().Query(
		Bool(Filter(filters...)),
	).Add(
		Aggs(
			AggItem{
				"body", Item{
					"terms": Item{
						"field": "network.keyword",
					},
				},
			},
		),
	).Zero()

	return e.GetCountAgg([]string{models.DocOperations}, query)
}

type getContractStatsByNetworkStats struct {
	Agg struct {
		Network struct {
			Buckets []struct {
				Bucket
				Same    IntValue   `json:"same"`
				Balance FloatValue `json:"balance"`
			} `json:"buckets"`
		} `json:"network"`
	} `json:"aggregations"`
}

// GetContractStatsByNetwork -
func (e *Elastic) GetContractStatsByNetwork(network string) (map[string]models.ContractCountStats, error) {
	query := NewQuery()

	if network != "" {
		query.Query(Bool(Filter(Match("network", network))))
	}

	query.Add(
		Aggs(
			AggItem{
				"network", Item{
					"terms": Item{
						"field": "network.keyword",
					},
					"aggs": Item{
						"same": Item{
							"cardinality": Item{
								"script": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value",
							},
						},
						"balance": Sum("balance"),
					},
				},
			},
		),
	).Zero()

	var response getContractStatsByNetworkStats
	if err := e.Query([]string{models.DocContracts}, query, &response); err != nil {
		return nil, err
	}

	counts := make(map[string]models.ContractCountStats)
	for _, item := range response.Agg.Network.Buckets {
		counts[item.Key] = models.ContractCountStats{
			Total:     item.DocCount,
			SameCount: item.Same.Value,
			Balance:   int64(item.Balance.Value),
		}
	}
	return counts, nil
}

// GetFACountByNetwork -
func (e *Elastic) GetFACountByNetwork(network string) (map[string]int64, error) {
	filters := []Item{
		In("tags", []string{
			"fa1",
			"fa12",
		}),
	}
	if network != "" {
		filters = append(filters, Match("network", network))
	}

	query := NewQuery().Query(
		Bool(Filter(filters...)),
	).Add(
		Aggs(
			AggItem{
				"body", Item{
					"terms": Item{
						"field": "network.keyword",
					},
				},
			},
		),
	).Zero()

	return e.GetCountAgg([]string{models.DocContracts}, query)
}

// GetLanguagesForNetwork -
func (e *Elastic) GetLanguagesForNetwork(network string) (map[string]int64, error) {
	query := NewQuery().Query(
		Bool(
			Filter(
				Match("network", network),
			),
		),
	).Add(
		Aggs(
			AggItem{
				"body", Item{
					"terms": Item{
						"field": "language.keyword",
					},
				},
			},
		),
	).Zero()

	return e.GetCountAgg([]string{models.DocContracts}, query)
}
