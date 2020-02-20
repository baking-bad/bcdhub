package elastic

import (
	"fmt"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

func parseContracts(res gjson.Result) []models.Contract {
	contracts := make([]models.Contract, 0)
	arr := res.Get("hits.hits").Array()
	for i := range arr {
		var c models.Contract
		c.ParseElasticJSON(arr[i])
		contracts = append(contracts, c)
	}
	return contracts
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
	res, err := e.query(DocContracts, q)
	if err != nil {
		return
	}
	if res.Get("hits.total.value").Int() < 1 {
		return c, fmt.Errorf("Unknown contract: %v", q)
	}
	hit := res.Get("hits.hits.0")
	c.ParseElasticJSON(hit)
	return
}

func (e *Elastic) getContracts(q map[string]interface{}) ([]models.Contract, error) {
	res, err := e.query(DocContracts, q)
	if err != nil {
		return nil, err
	}

	contracts := make([]models.Contract, 0)
	arr := res.Get("hits.hits").Array()
	for i := range arr {
		var c models.Contract
		c.ParseElasticJSON(arr[i])
		contracts = append(contracts, c)
	}
	return contracts, nil
}

// GetContract -
func (e *Elastic) GetContract(by map[string]interface{}) (models.Contract, error) {
	query := getContractQuery(by).One()
	return e.getContract(query)
}

// GetContractByID -
func (e *Elastic) GetContractByID(id string) (c models.Contract, err error) {
	resp, err := e.GetByID(DocContracts, id)
	if err != nil {
		return
	}
	if !resp.Get("found").Bool() {
		return c, fmt.Errorf("Unknown contract with ID %s", id)
	}
	c.ParseElasticJSON(resp)
	return
}

// GetContractField -
func (e *Elastic) GetContractField(by map[string]interface{}, field string) (interface{}, error) {
	query := getContractQuery(by).One()
	res, err := e.query(DocContracts, query, field)
	if err != nil {
		return nil, err
	}
	if res.Get("hits.total.value").Int() < 1 {
		return nil, fmt.Errorf("Unknown contract: %v", by)
	}
	return res.Get("hits.hits.0._source").Get(field).Value(), nil
}

// GetContracts -
func (e *Elastic) GetContracts(by map[string]interface{}) ([]models.Contract, error) {
	query := getContractQuery(by).All()
	return e.getContracts(query)
}

// GetRandomContract -
func (e *Elastic) GetRandomContract() (models.Contract, error) {
	query := newQuery().Query(qItem{
		"function_score": qItem{
			"functions": []qItem{
				qItem{
					"random_score": qItem{
						"seed": time.Now().UnixNano(),
					},
				},
			},
		},
	}).One()
	return e.getContract(query)
}

// GetContractStats -
func (e *Elastic) GetContractStats(address, network string) (stats ContractStats, err error) {
	b := boolQ(
		must(
			matchPhrase("network", network),
		),
		should(
			matchPhrase("source", address),
			matchPhrase("destination", address),
		),
	)
	b.Get("bool").Append("minimum_should_match", 1)
	query := newQuery().Query(b).Add(
		qItem{
			"aggs": qItem{
				"last_action":   max("timestamp"),
				"tx_count":      count("level"),
				"sum_tx_amount": sum("amount"),
			},
		},
	).Zero()
	res, err := e.query(DocOperations, query)
	if err != nil {
		return
	}
	stats.parse(res.Get("aggregations"))
	return
}
