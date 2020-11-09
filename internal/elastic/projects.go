package elastic

import (
	"math/rand"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

type getProjectsResponse struct {
	Agg struct {
		Projects struct {
			Buckets []struct {
				Bucket
				Last struct {
					Hits HitsArray `json:"hits"`
				} `json:"last"`
			} `json:"buckets"`
		} `json:"projects"`
	} `json:"aggregations"`
}

// GetProjectsLastContract -
func (e *Elastic) GetProjectsLastContract() ([]models.Contract, error) {
	query := newQuery().Add(
		aggs(
			aggItem{
				"projects", qItem{
					"terms": qItem{
						"field": "project_id.keyword",
						"size":  maxQuerySize,
					},
					"aggs": qItem{
						"last": topHits(1, "timestamp", "desc"),
					},
				},
			},
		),
	).Sort("timestamp", "desc").Zero()

	var response getProjectsResponse
	if err := e.query([]string{DocContracts}, query, &response); err != nil {
		return nil, err
	}

	if len(response.Agg.Projects.Buckets) == 0 {
		return nil, NewRecordNotFoundError(DocContracts, "", query)
	}

	contracts := make([]models.Contract, len(response.Agg.Projects.Buckets))
	for i := range response.Agg.Projects.Buckets {
		if err := json.Unmarshal(response.Agg.Projects.Buckets[i].Last.Hits.Hits[0].Source, &contracts[i]); err != nil {
			return nil, err
		}
	}
	return contracts, nil
}

// GetSameContracts -
func (e *Elastic) GetSameContracts(c models.Contract, size, offset int64) (pcr SameContractsResponse, err error) {
	if c.Fingerprint == nil {
		return pcr, errors.Errorf("Invalid contract data")
	}

	if size == 0 {
		size = defaultSize
	} else if size+offset > maxQuerySize {
		size = maxQuerySize - offset
	}

	q := newQuery().Query(
		boolQ(
			filter(
				matchPhrase("hash", c.Hash),
			),
			notMust(
				matchPhrase("address", c.Address),
			),
		),
	).Sort("last_action", "desc").Size(size).From(offset)

	var response SearchResponse
	if err = e.query([]string{DocContracts}, q, &response); err != nil {
		return
	}

	if len(response.Hits.Hits) == 0 {
		return pcr, NewRecordNotFoundError(DocContracts, "", q)
	}

	contracts := make([]models.Contract, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err = json.Unmarshal(response.Hits.Hits[i].Source, &contracts[i]); err != nil {
			return
		}
	}
	pcr.Contracts = contracts
	pcr.Count = response.Hits.Total.Value
	return
}

// GetSimilarContracts -
func (e *Elastic) GetSimilarContracts(c models.Contract, size, offset int64) (pcr []SimilarContract, total int, err error) {
	if c.Fingerprint == nil {
		return
	}

	if size == 0 {
		size = defaultSize
	} else if size+offset > maxQuerySize {
		size = maxQuerySize - offset
	}

	query := newQuery().Query(
		boolQ(
			filter(
				matchPhrase("project_id", c.ProjectID),
			),
			notMust(
				matchQ("hash.keyword", c.Hash),
			),
		),
	).Add(
		aggs(
			aggItem{
				"projects",
				qItem{
					"terms": qItem{
						"field": "hash.keyword",
						"size":  size + offset,
						"order": qItem{
							"bucketsSort": "desc",
						},
					},
					"aggs": qItem{
						"last":        topHits(1, "last_action", "desc"),
						"bucketsSort": max("last_action"),
					},
				},
			},
		),
	).Zero()

	var response getProjectsResponse
	if err = e.query([]string{DocContracts}, query, &response); err != nil {
		return
	}

	total = len(response.Agg.Projects.Buckets)
	if len(response.Agg.Projects.Buckets) == 0 {
		return
	}

	contracts := make([]SimilarContract, 0)
	arr := response.Agg.Projects.Buckets[offset:]
	for _, item := range arr {
		var contract models.Contract
		if err = json.Unmarshal(item.Last.Hits.Hits[0].Source, &contract); err != nil {
			return
		}

		similar := SimilarContract{
			Contract: &contract,
			Count:    item.DocCount,
		}
		contracts = append(contracts, similar)
	}
	return contracts, total, nil
}

type getDiffTasksResponse struct {
	Agg struct {
		Projects struct {
			Buckets []struct {
				Bucket
				Last struct {
					Hits HitsArray `json:"hits"`
				} `json:"last"`
				ByHash struct {
					Buckets []struct {
						Bucket
						Last struct {
							Hits HitsArray `json:"hits"`
						} `json:"last"`
					} `json:"buckets"`
				} `json:"by_hash"`
			} `json:"buckets"`
		} `json:"by_project"`
	} `json:"aggregations"`
}

// GetDiffTasks -
func (e *Elastic) GetDiffTasks() ([]DiffTask, error) {
	query := newQuery().Add(
		aggs(
			aggItem{
				"by_project", qItem{
					"terms": qItem{
						"field": "project_id.keyword",
						"size":  maxQuerySize,
					},
					"aggs": qItem{
						"by_hash": qItem{
							"terms": qItem{
								"field": "hash.keyword",
								"size":  maxQuerySize,
							},
							"aggs": qItem{
								"last": topHits(1, "last_action", "desc"),
							},
						},
					},
				},
			},
		),
	).Zero()

	var response getDiffTasksResponse
	if err := e.query([]string{DocContracts}, query, &response); err != nil {
		return nil, err
	}

	tasks := make([]DiffTask, 0)
	for _, bucket := range response.Agg.Projects.Buckets {
		if len(bucket.ByHash.Buckets) < 2 {
			continue
		}

		similar := bucket.ByHash.Buckets
		for i := 0; i < len(similar)-1; i++ {
			var current models.Contract
			if err := json.Unmarshal(similar[i].Last.Hits.Hits[0].Source, &current); err != nil {
				return nil, err
			}
			for j := i + 1; j < len(similar); j++ {
				var next models.Contract
				if err := json.Unmarshal(similar[j].Last.Hits.Hits[0].Source, &next); err != nil {
					return nil, err
				}

				tasks = append(tasks, DiffTask{
					Network1: current.Network,
					Address1: current.Address,
					Network2: next.Network,
					Address2: next.Address,
				})
			}
		}
	}

	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(tasks), func(i, j int) { tasks[i], tasks[j] = tasks[j], tasks[i] })
	return tasks, nil
}
