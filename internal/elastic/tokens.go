package elastic

import (
	"github.com/baking-bad/bcdhub/internal/models"
)

// GetTokens -
func (e *Elastic) GetTokens(network string, size, offset int64) ([]models.Contract, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				in("tags", []string{"fa12", "fa1"}),
			),
		),
	).Sort("timestamp", "desc").Size(size).From(offset)

	result, err := e.query([]string{DocContracts}, query)
	if err != nil {
		return nil, err
	}

	contracts := make([]models.Contract, 0)
	for _, hit := range result.Get("hits.hits").Array() {
		var contract models.Contract
		contract.ParseElasticJSON(hit)
		contracts = append(contracts, contract)
	}
	return contracts, nil
}

// GetTokenTransferOperations -
func (e *Elastic) GetTokenTransferOperations(network, address, lastID string, size int64) (PageableOperations, error) {
	if size == 0 {
		size = defaultSize
	}
	filterItems := []qItem{
		in("entrypoint", []string{"mint", "transfer"}),
		matchQ("parameter_strings", address),
		matchQ("network", network),
	}
	if lastID != "" {
		filterItems = append(filterItems, rangeQ("indexed_time", qItem{"lt": lastID}))
	}

	query := newQuery().Query(
		boolQ(
			filter(
				filterItems...,
			),
		),
	).Sort("timestamp", "desc").Size(size)

	po := PageableOperations{}
	result, err := e.query([]string{DocOperations}, query)
	if err != nil {
		return po, err
	}

	hits := result.Get("hits.hits").Array()
	operations := make([]models.Operation, len(hits))
	for i, hit := range hits {
		operations[i].ParseElasticJSON(hit)
		if i == len(hits)-1 {
			po.LastID = hit.Get("_source.indexed_time").String()
		}
	}
	po.Operations = operations
	return po, nil
}
