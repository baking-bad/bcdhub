package elastic

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

func parseContarctFromHit(hit gjson.Result, c *models.Contract) {
	c.ID = hit.Get("_id").String()
	c.Network = hit.Get("_source.network").String()
	c.Level = hit.Get("_source.level").Int()
	c.Timestamp = hit.Get("_source.timestamp").Time().UTC()
	c.Balance = hit.Get("_source.balance").Int()
	c.Kind = hit.Get("_source.kind").String()
	c.HashCode = hit.Get("_source.hash_code").String()
	c.Language = hit.Get("_source.language").String()

	c.Tags = make([]string, 0)
	for _, t := range hit.Get("_source.tags").Array() {
		c.Tags = append(c.Tags, t.String())
	}

	c.Hardcoded = make([]string, 0)
	for _, t := range hit.Get("_source.hardcoded").Array() {
		c.Hardcoded = append(c.Hardcoded, t.String())
	}

	c.Address = hit.Get("_source.address").String()
	c.Manager = hit.Get("_source.manager").String()
	c.Delegate = hit.Get("_source.delegate").String()
}

func getContractQuery(by map[string]interface{}) map[string]interface{} {
	match := []map[string]interface{}{}
	for k, v := range by {
		match = append(match, map[string]interface{}{
			"match": map[string]interface{}{
				k: v,
			},
		})
	}
	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": match,
			},
		},
	}
}

func (e *Elastic) getContract(q map[string]interface{}) (c models.Contract, err error) {
	res, err := e.query(DocContracts, q)
	if err != nil {
		return
	}
	if res.Get("hits.total.value").Int() != 1 {
		return c, fmt.Errorf("Unknown contract: %v", q)
	}
	hit := res.Get("hits.hits.0")
	parseContarctFromHit(hit, &c)
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
		parseContarctFromHit(arr[i], &c)
		contracts = append(contracts, c)
	}
	return contracts, nil
}

// GetContract -
func (e *Elastic) GetContract(by map[string]interface{}) (models.Contract, error) {
	query := getContractQuery(by)
	query["_source"] = map[string]interface{}{
		"excludes": []string{"hash_code"},
	}
	return e.getContract(query)
}

// GetContractField -
func (e *Elastic) GetContractField(by map[string]interface{}, field string) (interface{}, error) {
	query := getContractQuery(by)
	res, err := e.query(DocContracts, query, field)
	if err != nil {
		return nil, err
	}
	if res.Get("hits.total.value").Int() != 1 {
		return nil, fmt.Errorf("Unknown contract: %v", by)
	}
	return res.Get("hits.hits.0._source").Get(field).Value(), nil
}

// FindProjectContracts -
func (e *Elastic) FindProjectContracts(hashCode string, minScore float64) ([]models.Contract, error) {
	query := map[string]interface{}{
		"min_score": minScore,
		"size":      100,
		"_source": map[string]interface{}{
			"excludes": []string{"hash_code"},
		},
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"match": map[string]interface{}{
						"hash_code": hashCode,
					},
				},
			},
		},
	}
	return e.getContracts(query)
}

// SearchByText -
func (e *Elastic) SearchByText(text string) ([]models.Contract, error) {
	query := map[string]interface{}{
		"_source": map[string]interface{}{
			"excludes": []string{"hash_code"},
		},
		"size": 10,
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query": fmt.Sprintf("*%s*", text),
				"fields": []string{
					"address^10", "manager^8", "delegate^6", "tags^4", "hardcoded",
				},
			},
		},
	}
	return e.getContracts(query)
}
