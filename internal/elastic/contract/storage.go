package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/pkg/errors"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

func (storage *Storage) getContract(q core.Base) (c contract.Contract, err error) {
	var response core.SearchResponse
	if err = storage.es.Query([]string{models.DocContracts}, q, &response); err != nil {
		return
	}
	if response.Hits.Total.Value == 0 {
		return c, core.NewRecordNotFoundError(models.DocContracts, "")
	}
	err = json.Unmarshal(response.Hits.Hits[0].Source, &c)
	return
}

func (storage *Storage) getContracts(query core.Base) ([]contract.Contract, error) {
	contracts := make([]contract.Contract, 0)
	if err := storage.es.GetAllByQuery(query, &contracts); err != nil {
		return nil, err
	}

	return contracts, nil
}

// Get -
func (storage *Storage) Get(by map[string]interface{}) (contract.Contract, error) {
	query := core.FiltersToQuery(by).One()
	return storage.getContract(query)
}

// GetMany -
func (storage *Storage) GetMany(by map[string]interface{}) ([]contract.Contract, error) {
	query := core.FiltersToQuery(by)
	return storage.getContracts(query)
}

// GetRandom -
func (storage *Storage) GetRandom(network string) (contract.Contract, error) {
	random := core.Item{
		"function_score": core.Item{
			"functions": []core.Item{
				{
					"random_score": core.Item{
						"seed": time.Now().UnixNano(),
					},
				},
			},
		},
	}

	txRange := core.Range("tx_count", core.Item{
		"gte": 2,
	})

	must := []core.Item{txRange, random}
	if network != "" {
		must = append(must, core.Term("network", network))
	}

	query := core.NewQuery().Query(core.Bool(core.Must(must...))).One()

	return storage.getContract(query)
}

// IsFA -
func (storage *Storage) IsFA(network, address string) (bool, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Must(
				core.MatchPhrase("network", network),
				core.MatchPhrase("address", address),
			),
			core.Filter(
				core.Item{
					"terms": core.Item{
						"tags": []string{"fa12", "fa1"},
					},
				},
			),
		),
	)
	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocContracts}, query, &response, "address"); err != nil {
		return false, err
	}
	return response.Hits.Total.Value == 1, nil
}

// UpdateMigrationsCount -
func (storage *Storage) UpdateMigrationsCount(address, network string) error {
	// TODO: update via ID and script
	contract := contract.NewEmptyContract(network, address)
	if err := storage.es.GetByID(&contract); err != nil {
		return err
	}
	contract.MigrationsCount++
	return storage.es.UpdateDoc(&contract)
}

// GetAddressesByNetworkAndLevel -
func (storage *Storage) GetAddressesByNetworkAndLevel(network string, maxLevel int64) ([]string, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.Range("level", core.Item{
					"gt": maxLevel,
				}),
			),
		),
	).All()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocContracts}, query, &response, "address"); err != nil {
		return nil, err
	}

	addresses := make([]string, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		var c struct {
			Address string `json:"address"`
		}
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &c); err != nil {
			return nil, err
		}
		addresses[i] = c.Address
	}

	return addresses, nil
}

// GetIDsByAddresses -
func (storage *Storage) GetIDsByAddresses(addresses []string, network string) ([]string, error) {
	shouldItems := make([]core.Item, len(addresses))
	for i := range addresses {
		shouldItems[i] = core.MatchPhrase("address", addresses[i])
	}

	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
			),
			core.Should(shouldItems...),
			core.MinimumShouldMatch(1),
		),
	).Add(core.Item{
		"_source": false,
	})

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocContracts}, query, &response, "address"); err != nil {
		return nil, err
	}
	ids := make([]string, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		ids[i] = response.Hits.Hits[i].ID
	}
	return ids, nil
}

// GetByAddresses -
func (storage *Storage) GetByAddresses(addresses []contract.Address) ([]contract.Contract, error) {
	items := make([]core.Item, len(addresses))
	for i := range addresses {
		items[i] = core.Bool(
			core.Filter(
				core.MatchPhrase("address", addresses[i].Address),
				core.Match("network", addresses[i].Network),
			),
		)
	}

	query := core.NewQuery().Query(
		core.Bool(
			core.Should(items...),
			core.MinimumShouldMatch(1),
		),
	)
	contracts := make([]contract.Contract, 0)
	err := storage.es.GetAllByQuery(query, &contracts)
	return contracts, err
}

// GetProjectsLastContract -
func (storage *Storage) GetProjectsLastContract(c *contract.Contract) ([]contract.Contract, error) {
	query := core.NewQuery()

	if c != nil {
		filters := make([]core.Item, 0)
		if c.Manager != "" {
			filters = append(filters, core.MatchPhrase("manager", c.Manager))
		}
		if c.Language != "" {
			filters = append(filters, core.Term("language.keyword", c.Language))
		}
		if len(c.Tags) > 0 {
			filters = append(filters, getArrayFilter("tags", c.Tags...))
		}
		if len(c.Annotations) > 0 {
			filters = append(filters, getArrayFilter("annotations", c.Annotations...))
		}
		if len(c.FailStrings) > 0 {
			filters = append(filters, getArrayFilter("fail_strings", c.FailStrings...))
		}
		if len(c.Entrypoints) > 0 {
			filters = append(filters, getArrayFilter("entrypoints", c.Entrypoints...))
		}
		filters = append(filters, core.Bool(
			core.Must(
				core.Term("fingerprint.parameter.keyword", c.Fingerprint.Parameter),
				core.Term("fingerprint.storage.keyword", c.Fingerprint.Storage),
				core.Term("fingerprint.code.keyword", c.Fingerprint.Code),
			),
		))

		query.Query(
			core.Bool(
				core.Should(filters...),
				core.MinimumShouldMatch(1),
			),
		)
	}

	query.Add(
		core.Aggs(
			core.AggItem{
				Name: "projects",
				Body: core.Item{
					"terms": core.Item{
						"field": "project_id.keyword",
						"size":  core.MaxQuerySize,
					},
					"aggs": core.Item{
						"last": core.TopHits(1, "timestamp", "desc"),
					},
				},
			},
		),
	).Sort("timestamp", "desc").Zero()

	var response getProjectsResponse
	if err := storage.es.Query([]string{models.DocContracts}, query, &response); err != nil {
		return nil, err
	}

	if len(response.Agg.Projects.Buckets) == 0 {
		return nil, core.NewRecordNotFoundError(models.DocContracts, "")
	}

	contracts := make([]contract.Contract, len(response.Agg.Projects.Buckets))
	for i := range response.Agg.Projects.Buckets {
		if err := json.Unmarshal(response.Agg.Projects.Buckets[i].Last.Hits.Hits[0].Source, &contracts[i]); err != nil {
			return nil, err
		}
	}
	return contracts, nil
}

func getArrayFilter(fieldName string, arr ...string) core.Item {
	minimumShouldMatch := len(arr) / 2
	items := make([]core.Item, 0)
	for i := range arr {
		items = append(items, core.Term(fmt.Sprintf("%s.keyword", fieldName), arr[i]))
	}
	return core.Bool(
		core.Should(items...),
		core.MinimumShouldMatch(minimumShouldMatch),
	)
}

// GetSameContracts -
func (storage *Storage) GetSameContracts(c contract.Contract, manager string, size, offset int64) (pcr contract.SameResponse, err error) {
	if c.Fingerprint == nil {
		return pcr, errors.Errorf("Invalid contract data")
	}

	if size == 0 {
		size = consts.DefaultSize
	} else if size+offset > core.MaxQuerySize {
		size = core.MaxQuerySize - offset
	}

	var filter core.Item
	if manager == "" {
		filter = core.Filter(core.MatchPhrase("hash", c.Hash))
	} else {
		filter = core.Filter(
			core.MatchPhrase("hash", c.Hash),
			core.MatchPhrase("manager", manager),
		)
	}

	q := core.NewQuery().Query(
		core.Bool(
			filter,
			core.MustNot(
				core.MatchPhrase("address", c.Address),
			),
		),
	).Sort("last_action", "desc").Size(size).From(offset)

	var response core.SearchResponse
	if err = storage.es.Query([]string{models.DocContracts}, q, &response); err != nil {
		return
	}

	if len(response.Hits.Hits) == 0 {
		return pcr, core.NewRecordNotFoundError(models.DocContracts, "")
	}

	contracts := make([]contract.Contract, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err = json.Unmarshal(response.Hits.Hits[i].Source, &contracts[i]); err != nil {
			return
		}
	}
	pcr.Contracts = contracts
	if response.Hits.Total.Relation == "eq" {
		pcr.Count = response.Hits.Total.Value
	} else {
		countQuery := core.NewQuery().Query(
			core.Bool(
				filter,
				core.MustNot(
					core.MatchPhrase("address", c.Address),
				),
			),
		).Sort("last_action", "desc")
		pcr.Count, err = storage.es.CountItems([]string{models.DocContracts}, countQuery)
		if err != nil {
			return
		}
	}
	return
}

// GetSimilarContracts -
func (storage *Storage) GetSimilarContracts(c contract.Contract, size, offset int64) (pcr []contract.Similar, total int, err error) {
	if c.Fingerprint == nil {
		return
	}

	if size == 0 {
		size = consts.DefaultSize
	} else if size+offset > core.MaxQuerySize {
		size = core.MaxQuerySize - offset
	}

	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.MatchPhrase("project_id", c.ProjectID),
			),
			core.MustNot(
				core.Match("hash.keyword", c.Hash),
			),
		),
	).Add(
		core.Aggs(
			core.AggItem{
				Name: "projects",
				Body: core.Item{
					"terms": core.Item{
						"field": "hash.keyword",
						"size":  size + offset,
						"order": core.Item{
							"bucketsSort": "desc",
						},
					},
					"aggs": core.Item{
						"last":        core.TopHits(1, "last_action", "desc"),
						"bucketsSort": core.Max("last_action"),
					},
				},
			},
		),
	).Zero()

	var response getProjectsResponse
	if err = storage.es.Query([]string{models.DocContracts}, query, &response); err != nil {
		return
	}

	total = len(response.Agg.Projects.Buckets)
	if len(response.Agg.Projects.Buckets) == 0 {
		return
	}

	contracts := make([]contract.Similar, 0)
	arr := response.Agg.Projects.Buckets[offset:]
	for _, item := range arr {
		var cntr contract.Contract
		if err = json.Unmarshal(item.Last.Hits.Hits[0].Source, &cntr); err != nil {
			return
		}

		contracts = append(contracts, contract.Similar{
			Contract: &cntr,
			Count:    item.DocCount,
		})
	}
	return contracts, total, nil
}

// GetDiffTasks -
func (storage *Storage) GetDiffTasks() ([]contract.DiffTask, error) {
	query := core.NewQuery().Add(
		core.Aggs(
			core.AggItem{
				Name: "by_project",
				Body: core.Item{
					"terms": core.Item{
						"field": "project_id.keyword",
						"size":  core.MaxQuerySize,
					},
					"aggs": core.Item{
						"by_hash": core.Item{
							"terms": core.Item{
								"field": "hash.keyword",
								"size":  core.MaxQuerySize,
							},
							"aggs": core.Item{
								"last": core.TopHits(1, "last_action", "desc"),
							},
						},
					},
				},
			},
		),
	).Zero()

	var response getDiffTasksResponse
	if err := storage.es.Query([]string{models.DocContracts}, query, &response); err != nil {
		return nil, err
	}

	tasks := make([]contract.DiffTask, 0)
	for _, bucket := range response.Agg.Projects.Buckets {
		if len(bucket.ByHash.Buckets) < 2 {
			continue
		}

		similar := bucket.ByHash.Buckets
		for i := 0; i < len(similar)-1; i++ {
			var current contract.Contract
			if err := json.Unmarshal(similar[i].Last.Hits.Hits[0].Source, &current); err != nil {
				return nil, err
			}
			for j := i + 1; j < len(similar); j++ {
				var next contract.Contract
				if err := json.Unmarshal(similar[j].Last.Hits.Hits[0].Source, &next); err != nil {
					return nil, err
				}

				tasks = append(tasks, contract.DiffTask{
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

// GetTokens -
func (storage *Storage) GetTokens(network, tokenInterface string, offset, size int64) ([]contract.Contract, int64, error) {
	tags := []string{"fa12", "fa1", "fa2"}
	if tokenInterface == "fa12" || tokenInterface == "fa1" || tokenInterface == "fa2" {
		tags = []string{tokenInterface}
	}

	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.In("tags", tags),
			),
		),
	).Sort("timestamp", "desc")

	contracts := make([]contract.Contract, 0)
	ctx := core.NewScrollContext(storage.es, query, size, consts.DefaultScrollSize)
	ctx.Offset = offset
	if err := ctx.Get(&contracts); err != nil {
		return nil, 0, err
	}

	countQuery := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.In("tags", tags),
			),
		),
	)
	count, err := storage.es.CountItems([]string{models.DocContracts}, countQuery)
	if err != nil {
		return nil, 0, err
	}

	return contracts, count, nil
}

// UpdateField -
func (storage *Storage) UpdateField(where []contract.Contract, fields ...string) error {
	if len(where) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range where {
		updated, err := storage.es.BuildFieldsForModel(where[i], fields...)
		if err != nil {
			return err
		}
		meta := fmt.Sprintf(`{ "update": { "_id": "%s", "_index": "%s", "retry_on_conflict": 2}}%s%s%s`, where[i].GetID(), where[i].GetIndex(), "\n", string(updated), "\n")
		bulk.Grow(len(meta))
		bulk.WriteString(meta)

		if (i%1000 == 0 && i > 0) || i == len(where)-1 {
			if err := storage.es.Bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}
