package elastic

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

type getCountAggResponse struct {
	Agg struct {
		Body struct {
			Buckets []Bucket `json:"buckets"`
		} `json:"body"`
	} `json:"aggregations"`
}

// GetCountAgg -
func (e *Elastic) GetCountAgg(index []string, query Base) (map[string]int64, error) {
	var response getCountAggResponse
	if err := e.query(index, query, &response); err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, item := range response.Agg.Body.Buckets {
		counts[item.Key] = int64(item.DocCount)
	}
	return counts, nil
}

// NetworkCountStats -
func (e *Elastic) NetworkCountStats(network types.Network) (map[string]int64, error) {
	query := NewQuery().Query(
		Bool(
			Filter(
				Match("network", network),
			),
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

type getContractStatsByNetworkStats struct {
	Agg struct {
		Network struct {
			Buckets []struct {
				Bucket
				Same UintValue `json:"same"`
			} `json:"buckets"`
		} `json:"network"`
	} `json:"aggregations"`
}

// NetworkStats -
func (e *Elastic) NetworkStats(network types.Network) (map[string]*models.NetworkStats, error) {
	query := NewQuery()

	if network != types.Empty {
		query.Query(Bool(Filter(Match("network", network.String()))))
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
								"script": "doc['hash.keyword'].value",
							},
						},
					},
				},
			},
		),
	).Zero()

	var response getContractStatsByNetworkStats
	if err := e.query([]string{models.DocContracts}, query, &response); err != nil {
		return nil, err
	}

	counts := make(map[string]*models.NetworkStats)
	for _, item := range response.Agg.Network.Buckets {
		counts[item.Key] = &models.NetworkStats{
			ContractsCount:       item.DocCount,
			UniqueContractsCount: item.Same.Value,
			CallsCount:           0,
		}
	}

	faCount, err := e.getFACountByNetwork(network)
	if err != nil {
		return nil, err
	}

	for network, count := range faCount {
		if stats, ok := counts[network]; ok {
			stats.FACount = uint64(count)
		} else {
			counts[network] = &models.NetworkStats{
				FACount: uint64(count),
			}
		}
	}

	callsCount, err := e.callsCount(network)
	if err != nil {
		return nil, err
	}

	for network, count := range callsCount {
		if stats, ok := counts[network]; ok {
			stats.CallsCount = uint64(count)
		} else {
			counts[network] = &models.NetworkStats{
				CallsCount: uint64(count),
			}
		}
	}

	return counts, nil
}

// LanguageByNetwork -
func (e *Elastic) LanguageByNetwork(network types.Network) (map[string]int64, error) {
	query := NewQuery().Query(
		Bool(
			Filter(
				Match("network", network.String()),
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

func (e *Elastic) getFACountByNetwork(network types.Network) (map[string]int64, error) {
	filters := []Item{
		In("tags", []string{
			"fa1",
			"fa12",
		}),
	}
	if network != types.Empty {
		filters = append(filters, Match("network", network.String()))
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

func (e *Elastic) callsCount(network types.Network) (map[string]int64, error) {
	filters := []Item{
		Exists("entrypoint"),
	}
	if network != types.Empty {
		filters = append(filters, Match("network", network.String()))
	}

	query := NewQuery().Query(
		Bool(
			Filter(filters...),
			MustNot(Match("entrypoint", "")),
		),
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
