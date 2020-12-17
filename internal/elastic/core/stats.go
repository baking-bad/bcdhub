package core

import (
	"github.com/baking-bad/bcdhub/internal/elastic/consts"
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

	return e.GetCountAgg([]string{consts.DocContracts, consts.DocOperations}, query)
}

// GetCallsCountByNetwork -
func (e *Elastic) GetCallsCountByNetwork() (map[string]int64, error) {
	query := NewQuery().Query(Exists("entrypoint")).Add(
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

	return e.GetCountAgg([]string{consts.DocOperations}, query)
}

type getContractStatsByNetworkStats struct {
	Agg struct {
		Network struct {
			Buckets []struct {
				Bucket
				Same           IntValue   `json:"same"`
				Balance        FloatValue `json:"balance"`
				TotalWithdrawn FloatValue `json:"total_withdrawn"`
			} `json:"buckets"`
		} `json:"network"`
	} `json:"aggregations"`
}

// GetContractStatsByNetwork -
func (e *Elastic) GetContractStatsByNetwork() (map[string]models.ContractCountStats, error) {
	query := NewQuery().Add(
		Aggs(
			AggItem{
				"network", Item{
					"terms": Item{
						"field": "network.keyword",
					},
					"core.Aggs": Item{
						"same": Item{
							"cardinality": Item{
								"script": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value",
							},
						},
						"balance":         Sum("balance"),
						"total_withdrawn": Sum("total_withdrawn"),
					},
				},
			},
		),
	).Zero()

	var response getContractStatsByNetworkStats
	if err := e.Query([]string{consts.DocContracts}, query, &response); err != nil {
		return nil, err
	}

	counts := make(map[string]models.ContractCountStats)
	for _, item := range response.Agg.Network.Buckets {
		counts[item.Key] = models.ContractCountStats{
			Total:          item.DocCount,
			SameCount:      item.Same.Value,
			Balance:        int64(item.Balance.Value),
			TotalWithdrawn: int64(item.TotalWithdrawn.Value),
		}
	}
	return counts, nil
}

// GetFACountByNetwork -
func (e *Elastic) GetFACountByNetwork() (map[string]int64, error) {
	query := NewQuery().Query(
		In("tags", []string{
			"fa1",
			"fa12",
		}),
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

	return e.GetCountAgg([]string{consts.DocContracts}, query)
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

	return e.GetCountAgg([]string{consts.DocContracts}, query)
}
