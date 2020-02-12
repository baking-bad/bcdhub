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
