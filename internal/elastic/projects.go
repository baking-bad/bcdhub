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
	query := newQuery().Add(
		aggs("projects", qItem{
			"terms": qItem{
				"field": "project_id.keyword",
				"size":  maxQuerySize,
			},
			"aggs": qItem{
				"last": topHits(1, "timestamp", "desc"),
			},
		}),
	).Zero()

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
		parseContractFromHit(item, &c)
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
		parseContractFromHit(item, &c)
		contracts = append(contracts, c)
	}
	return contracts, nil
}

// GetSimilarContracts -
func (e *Elastic) GetSimilarContracts(c models.Contract) ([]map[string]interface{}, error) {
	if c.Fingerprint == nil {
		return nil, nil
	}
	fgpt := fmt.Sprintf("%s|%s|%s", c.Fingerprint.Parameter, c.Fingerprint.Storage, c.Fingerprint.Code)

	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("project_id", c.ProjectID),
			),
			notMust(
				qItem{
					"script": qItem{
						"script": qItem{
							"source": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value == params.fgpt",
							"params": qItem{
								"fgpt": fgpt,
							},
						},
					},
				},
			),
		),
	).Add(
		aggs(
			"projects",
			qItem{
				"terms": qItem{
					"script": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value",
					"size":   10000,
					"order": qItem{
						"bucketsSort": "desc",
					},
				},
				"aggs": qItem{
					"last":        topHits(1, "timestamp", "desc"),
					"bucketsSort": max("timestamp"),
				},
			},
		),
	).Zero()

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
		parseContractFromHit(item.Get("last.hits.hits.0"), &c)
		res = append(res, qItem{
			"count": item.Get("doc_count").Int(),
			"last":  c,
		})
	}
	return res, nil
}

// GetProjectsStats -
func (e *Elastic) GetProjectsStats() (stats []ProjectStats, err error) {
	last := topHits(1, "timestamp", "desc")
	last.Get("top_hits").Append("_source", includes([]string{"address", "network", "timestamp"}))

	query := newQuery().Add(
		aggs("by_project", qItem{
			"terms": qItem{
				"field": "project_id.keyword",
				"size":  maxQuerySize,
			},
			"aggs": qItem{
				"by_same": qItem{
					"terms": qItem{
						"script": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value",
						"size":   maxQuerySize,
					},
					"aggs": qItem{
						"last_action_date":  max("last_action"),
						"first_deploy_date": min("timestamp"),
					},
				},
				"count": qItem{
					"cardinality": qItem{
						"script": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value",
					},
				},
				"last_action_date":  maxBucket("by_same>last_action_date"),
				"first_deploy_date": minBucket("by_same>first_deploy_date"),
				"language": qItem{
					"terms": qItem{
						"field": "language.keyword",
						"size":  1,
					},
				},
				"tx_count": sum("tx_count"),
				"last":     last,
			},
		}),
	).Zero()
	resp, err := e.query(DocContracts, query)
	if err != nil {
		return
	}
	count := resp.Get("aggregations.by_project.buckets.#").Int()
	stats = make([]ProjectStats, count)
	for i, item := range resp.Get("aggregations.by_project.buckets").Array() {
		var p ProjectStats
		p.parse(item)
		stats[i] = p
	}
	return
}

// GetProjectStats -
func (e *Elastic) GetProjectStats(projectID string) (p ProjectStats, err error) {
	last := topHits(1, "timestamp", "desc")
	last.Get("top_hits").Append("_source", includes([]string{"address", "network", "timestamp"}))

	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("project_id", projectID),
			),
		),
	).Add(
		qItem{
			"aggs": qItem{
				"by_same": qItem{
					"terms": qItem{
						"script": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value",
						"size":   maxQuerySize,
					},
					"aggs": qItem{
						"last_action_date":  max("last_action"),
						"first_deploy_date": min("timestamp"),
					},
				},
				"count": qItem{
					"cardinality": qItem{
						"script": "doc['fingerprint.parameter'].value + '|' + doc['fingerprint.storage'].value + '|' + doc['fingerprint.code'].value",
					},
				},
				"last_action_date":  maxBucket("by_same>last_action_date"),
				"first_deploy_date": minBucket("by_same>first_deploy_date"),
				"language": qItem{
					"terms": qItem{
						"field": "language.keyword",
						"size":  1,
					},
				},
				"tx_count": sum("tx_count"),
				"last":     last,
			},
		},
	).Zero()
	resp, err := e.query(DocContracts, query)
	if err != nil {
		return
	}
	p.parse(resp.Get("aggregations"))
	return
}
