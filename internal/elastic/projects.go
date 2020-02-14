package elastic

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

func parseProjectFormHit(hit gjson.Result, proj *models.Project) {
	proj.ID = hit.Get("_id").String()
	proj.Alias = hit.Get("_source.alias").String()
}

// GetProject -
func (e *Elastic) GetProject(id string) (p models.Project, err error) {
	res, err := e.GetByID(DocProjects, id)
	if err != nil {
		return
	}
	if !res.Get("found").Bool() {
		return p, fmt.Errorf("Unknown project: %s", id)
	}
	parseProjectFormHit(*res, &p)
	return
}

// GetLastProjectContracts -
func (e *Elastic) GetLastProjectContracts() ([]models.Contract, error) {
	query := map[string]interface{}{
		"size": 0,
		"aggs": map[string]interface{}{
			"projects": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "project_id.keyword",
					"size":  10000,
				},
				"aggs": map[string]interface{}{
					"last": map[string]interface{}{
						"top_hits": map[string]interface{}{
							"size": 1,
							"sort": map[string]interface{}{
								"timestamp": map[string]interface{}{
									"order": "desc",
								},
							},
						},
					},
				},
			},
		},
	}

	resp, err := e.query(DocContracts, query)
	if err != nil {
		return nil, err
	}

	arr := resp.Get("aggregations.projects.buckets.#.last.hits.hits.0")
	if !arr.Exists() {
		return nil, fmt.Errorf("Empty response: %v", resp)
	}

	contracts := make([]models.Contract, 0)
	for _, item := range arr.Array() {
		var c models.Contract
		parseContarctFromHit(item, &c)
		contracts = append(contracts, c)
	}
	return contracts, nil
}

// GetSameContracts -
func (e *Elastic) GetSameContracts(c models.Contract) ([]models.Contract, error) {
	if c.Fingerprint == nil {
		return nil, fmt.Errorf("Invalid contract data")
	}

	q := newQuery().Query(
		boolQ(
			must(
				matchPhrase("fingerprint.parameter", c.Fingerprint.Parameter),
				matchPhrase("fingerprint.storage", c.Fingerprint.Storage),
				matchPhrase("fingerprint.code", c.Fingerprint.Code),
			),
		),
	).Sort("timestamp", "desc").All()

	resp, err := e.query(DocContracts, q)
	if err != nil {
		return nil, err
	}

	if resp.Get("hits.total.value").Int() < 1 {
		return nil, fmt.Errorf("Unknown contract: %v", c.Address)
	}

	arr := resp.Get("hits.hits")
	if !arr.Exists() {
		return nil, fmt.Errorf("Empty response: %v", resp)
	}

	contracts := make([]models.Contract, 0)
	for _, item := range arr.Array() {
		var c models.Contract
		parseContarctFromHit(item, &c)
		contracts = append(contracts, c)
	}
	return contracts, nil
}

// GetSimilarContracts -
func (e *Elastic) GetSimilarContracts(c models.Contract) ([]map[string]interface{}, error) {
	if c.ProjectID == "" || c.Fingerprint == nil {
		return nil, fmt.Errorf("Invalid contract data")
	}

	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"match_phrase": map[string]interface{}{
						"project_id": c.ProjectID,
					},
				},
				"must_not": []map[string]interface{}{
					map[string]interface{}{
						"match_phrase": map[string]interface{}{
							"fingerprint.parameter": c.Fingerprint.Parameter,
						},
					},
					map[string]interface{}{
						"match_phrase": map[string]interface{}{
							"fingerprint.storage": c.Fingerprint.Parameter,
						},
					},
					map[string]interface{}{
						"match_phrase": map[string]interface{}{
							"fingerprint.code": c.Fingerprint.Parameter,
						},
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"projects": map[string]interface{}{
				"terms": map[string]interface{}{
					"script": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value",
					"size":   10000,
					"order": map[string]interface{}{
						"bucketsSort": "desc",
					},
				},
				"aggs": map[string]interface{}{
					"last": map[string]interface{}{
						"top_hits": map[string]interface{}{
							"size": 1,
							"sort": map[string]interface{}{
								"timestamp": map[string]interface{}{
									"order": "desc",
								},
							},
						},
					},
					"bucketsSort": map[string]interface{}{
						"max": map[string]interface{}{
							"field": "timestamp",
						},
					},
				},
			},
		},
	}

	resp, err := e.query(DocContracts, query)
	if err != nil {
		return nil, err
	}

	buckets := resp.Get("aggregations.projects.buckets")
	if !buckets.Exists() {
		return nil, nil
	}

	res := make([]map[string]interface{}, 0)
	for _, item := range buckets.Array() {
		var c models.Contract
		parseContarctFromHit(item.Get("last.hits.hits.0"), &c)
		res = append(res, qItem{
			"count": item.Get("doc_count").Int(),
			"last": c,
		})
	}
	return res, nil
}
