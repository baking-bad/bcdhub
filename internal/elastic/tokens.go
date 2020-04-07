package elastic

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
)

// GetTokens -
func (e *Elastic) GetTokens(network string, size, offset int64) ([]models.Contract, error) {
	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("network", network),
			),
			filter(
				qItem{
					"terms": qItem{
						"tags": []string{"fa12", "fa1"},
					},
				},
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
func (e *Elastic) GetTokenTransferOperations(network, address, lastID string) (PageableOperations, error) {
	mustItems := []qItem{
		matchPhrase("network", network),
		matchPhrase("entrypoint", "transfer"),
		matchQ("parameters", fmt.Sprintf(".*%s.*", address)),
	}
	if lastID != "" {
		mustItems = append(mustItems, rangeQ("indexed_time", qItem{"lt": lastID}))
	}

	query := newQuery().Query(
		boolQ(must(mustItems...)),
	).Add(
		aggs("last_id", min("indexed_time")),
	).Sort("timestamp", "desc").Size(20)

	po := PageableOperations{}
	result, err := e.query([]string{DocOperations}, query)
	if err != nil {
		return po, err
	}

	operations := make([]models.Operation, 0)
	for _, hit := range result.Get("hits.hits").Array() {
		var operation models.Operation
		operation.ParseElasticJSON(hit)
		operations = append(operations, operation)
	}
	po.Operations = operations
	po.LastID = result.Get("aggregations.last_id.value").String()
	return po, nil
}

// GetTokenBalance -
func (e *Elastic) GetTokenBalance(network, address, tokenAddress string) (int64, error) {
	return 0, nil
}

// GetAllTokenBalances -
func (e *Elastic) GetAllTokenBalances(network, tokenAddress string) (map[string]int64, error) {
	return nil, nil
}

// GetTokenTotalSupply -
func (e *Elastic) GetTokenTotalSupply(network, tokenAddress string) (int64, error) {
	return 0, nil
}

// GetTokenMinted -
func (e *Elastic) GetTokenMinted(network, tokenAddress string) (int64, error) {
	return 0, nil
}

// GetTokenBurned -
func (e *Elastic) GetTokenBurned(network, tokenAddress string) (int64, error) {
	return 0, nil
}
