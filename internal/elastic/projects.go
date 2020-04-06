package elastic

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
)

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

	resp, err := e.query([]string{DocContracts}, query)
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
		c.ParseElasticJSON(item)
		contracts = append(contracts, c)
	}
	return contracts, nil
}

// GetSameContracts -
func (e *Elastic) GetSameContracts(c models.Contract, size, offset int64) (scp SameContractsResponse, err error) {
	if c.Fingerprint == nil {
		return scp, fmt.Errorf("Invalid contract data")
	}

	if size == 0 {
		size = 10
	}

	q := newQuery().Query(
		boolQ(
			must(
				matchPhrase("fingerprint.parameter", c.Fingerprint.Parameter),
				matchPhrase("fingerprint.storage", c.Fingerprint.Storage),
				matchPhrase("fingerprint.code", c.Fingerprint.Code),
			),
			notMust(
				matchPhrase("address", c.Address),
			),
		),
	).Add(qItem{
		"sort": qItem{
			"_script": qItem{
				"type": "number",
				"script": qItem{
					"lang":   "painless",
					"inline": "doc['network.keyword'].value == 'mainnet' ? (2 * doc['timestamp'].value.toInstant().toEpochMilli()) : doc['timestamp'].value.toInstant().toEpochMilli()",
				},
				"order": "desc",
			},
		},
	}).Size(size).From(offset)

	resp, err := e.query([]string{DocContracts}, q)
	if err != nil {
		return
	}

	arr := resp.Get("hits.hits")
	if !arr.Exists() {
		return scp, fmt.Errorf("Empty response: %v", resp)
	}

	contracts := make([]models.Contract, 0)
	for _, item := range arr.Array() {
		var c models.Contract
		c.ParseElasticJSON(item)
		contracts = append(contracts, c)
	}
	scp.Contracts = contracts
	scp.Count = resp.Get("hits.total.value").Uint()
	return
}

// GetSimilarContracts -
func (e *Elastic) GetSimilarContracts(c models.Contract) ([]SimilarContract, error) {
	if c.Fingerprint == nil {
		return nil, nil
	}
	fgpt := fmt.Sprintf("%s|%s|%s", c.Fingerprint.Parameter, c.Fingerprint.Storage, c.Fingerprint.Code)

	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("project_id", c.ProjectID),
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

	resp, err := e.query([]string{DocContracts}, query)
	if err != nil {
		return nil, err
	}

	buckets := resp.Get("aggregations.projects.buckets")
	if !buckets.Exists() {
		return nil, nil
	}

	res := make([]SimilarContract, 0)
	for _, item := range buckets.Array() {
		var buf models.Contract
		buf.ParseElasticJSON(item.Get("last.hits.hits.0"))
		bufF := fmt.Sprintf("%s|%s|%s", buf.Fingerprint.Parameter, buf.Fingerprint.Storage, buf.Fingerprint.Code)
		if fgpt == bufF {
			continue
		}
		res = append(res, SimilarContract{
			Contract: &buf,
			Count:    item.Get("doc_count").Int(),
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
	resp, err := e.query([]string{DocContracts}, query)
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
	resp, err := e.query([]string{DocContracts}, query)
	if err != nil {
		return
	}
	p.parse(resp.Get("aggregations"))
	return
}

func (e *Elastic) getProjectsContracts(ids []string) (res []contractPair, err error) {
	if len(ids) == 0 {
		return
	}

	query := "SELECT address, network FROM contract WHERE project_id IN (%s)"

	inString := "("
	for i := range ids {
		inString += fmt.Sprintf("'%s'", ids[i])
		if i != len(ids)-1 {
			inString += ","
		}
	}
	inString += ")"

	query = fmt.Sprintf(query, inString)
	resp, err := e.executeSQL(query)
	if err != nil {
		return
	}
	res = make([]contractPair, 0)
	for _, hit := range resp.Get("rows").Array() {
		var cid contractPair
		cid.ParseElasticJSONArray(hit)
		res = append(res, cid)
	}
	return res, err
}
