package core

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/restream/reindexer"
)

func countByField(field string, query *reindexer.Query) (map[string]int64, error) {
	query.AggregateFacet(field)

	it := query.Exec()
	defer it.Close()

	if it.Error() != nil {
		return nil, it.Error()
	}

	aggRes := it.AggResults()[0]

	response := make(map[string]int64)
	for i := range aggRes.Facets {
		response[aggRes.Facets[i].Values[0]] = int64(aggRes.Facets[i].Count)
	}
	return response, nil
}

// GetNetworkCountStats -
func (r *Reindexer) GetNetworkCountStats(network string) (map[string]int64, error) {
	res := make(map[string]int64)
	for _, index := range []string{models.DocContracts, models.DocOperations} {
		query := r.Query(index).Match("network", network)

		count, err := r.Count(query)
		if err != nil {
			return nil, err
		}
		res[index] = count
	}
	return res, nil
}

// GetCallsCountByNetwork -
func (r *Reindexer) GetCallsCountByNetwork() (map[string]int64, error) {
	query := r.Query(models.DocContracts).
		WhereString("entrypoint", reindexer.EMPTY, "")

	return countByField("network", query)
}

// GetContractStatsByNetwork -
// TODO: to do =)
func (r *Reindexer) GetContractStatsByNetwork() (map[string]models.ContractCountStats, error) {
	// query := NewQuery().Add(
	// 	Aggs(
	// 		AggItem{
	// 			"network", Item{
	// 				"terms": Item{
	// 					"field": "network.keyword",
	// 				},
	// 				"core.Aggs": Item{
	// 					"same": Item{
	// 						"cardinality": Item{
	// 							"script": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value",
	// 						},
	// 					},
	// 					"balance":         Sum("balance"),
	// 				},
	// 			},
	// 		},
	// 	),
	// ).Zero()

	// var response getContractStatsByNetworkStats
	// if err := e.Query([]string{models.DocContracts}, query, &response); err != nil {
	// 	return nil, err
	// }

	// counts := make(map[string]models.ContractCountStats)
	// for _, item := range response.Agg.Network.Buckets {
	// 	counts[item.Key] = models.ContractCountStats{
	// 		Total:          item.DocCount,
	// 		SameCount:      item.Same.Value,
	// 		Balance:        int64(item.Balance.Value),
	// 		TotalWithdrawn: int64(item.TotalWithdrawn.Value),
	// 	}
	// }
	// return counts, nil
	return nil, nil
}

// GetFACountByNetwork -
func (r *Reindexer) GetFACountByNetwork() (map[string]int64, error) {
	query := r.Query(models.DocContracts).Match("tags", "fa1", "fa12")
	return countByField("network", query)
}

// GetLanguagesForNetwork -
func (r *Reindexer) GetLanguagesForNetwork(network string) (map[string]int64, error) {
	query := r.Query(models.DocContracts).Match("network", network)
	return countByField("language", query)
}
