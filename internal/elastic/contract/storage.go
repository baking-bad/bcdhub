package contract

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/helpers"
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
	if err = storage.es.Query([]string{consts.DocContracts}, q, &response); err != nil {
		return
	}
	if response.Hits.Total.Value == 0 {
		return c, core.NewRecordNotFoundError(consts.DocContracts, "")
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
func (storage *Storage) GetRandom() (contract.Contract, error) {
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
	b := core.Bool(core.Must(txRange, random))
	query := core.NewQuery().Query(b).One()
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
	if err := storage.es.Query([]string{consts.DocContracts}, query, &response, "address"); err != nil {
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
	if err := storage.es.Query([]string{consts.DocContracts}, query, &response, "address"); err != nil {
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
	if err := storage.es.Query([]string{consts.DocContracts}, query, &response, "address"); err != nil {
		return nil, err
	}
	ids := make([]string, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		ids[i] = response.Hits.Hits[i].ID
	}
	return ids, nil
}

// RecalcStats -
func (storage *Storage) RecalcStats(network, address string) (stats contract.Stats, err error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
			),
			core.Should(
				core.MatchPhrase("source", address),
				core.MatchPhrase("destination", address),
			),
			core.MinimumShouldMatch(1),
		),
	).Add(
		core.Item{
			"aggs": core.Item{
				"tx_count":    core.Count("indexed_time"),
				"last_action": core.Max("timestamp"),
				"balance": core.Item{
					"scripted_metric": core.Item{
						"init_script":    "state.operations = []",
						"map_script":     "if (doc['status.keyword'].value == 'applied' && doc['amount'].size() != 0) {state.operations.add(doc['destination.keyword'].value == params.address ? doc['amount'].value : -1L * doc['amount'].value)}",
						"combine_script": "double balance = 0; for (amount in state.operations) { balance += amount } return balance",
						"reduce_script":  "double balance = 0; for (a in states) { balance += a } return balance",
						"params": core.Item{
							"address": address,
						},
					},
				},
				"total_withdrawn": core.Item{
					"scripted_metric": core.Item{
						"init_script":    "state.operations = []",
						"map_script":     "if (doc['status.keyword'].value == 'applied' && doc['amount'].size() != 0 && doc['source.keyword'].value == params.address) {state.operations.add(doc['amount'].value)}",
						"combine_script": "double balance = 0; for (amount in state.operations) { balance += amount } return balance",
						"reduce_script":  "double balance = 0; for (a in states) { balance += a } return balance",
						"params": core.Item{
							"address": address,
						},
					},
				},
			},
		},
	).Zero()
	var response recalcContractStatsResponse
	if err = storage.es.Query([]string{consts.DocOperations}, query, &response); err != nil {
		return
	}

	stats.LastAction = time.Unix(0, response.Aggs.LastAction.Value*1000000).UTC()
	stats.Balance = response.Aggs.Balance.Value
	stats.TotalWithdrawn = response.Aggs.TotalWithdrawn.Value
	stats.TxCount = response.Aggs.TxCount.Value
	return
}

// GetMigrationsCount -
func (storage *Storage) GetMigrationsCount(network, address string) (int64, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
			),
			core.Should(
				core.MatchPhrase("source", address),
				core.MatchPhrase("destination", address),
			),
			core.MinimumShouldMatch(1),
		),
	).Add(
		core.Aggs(
			core.AggItem{
				Name: "migrations_count",
				Body: core.Count("indexed_time"),
			},
		),
	).Zero()

	var response getContractMigrationStatsResponse
	err := storage.es.Query([]string{consts.DocMigrations}, query, &response)
	return response.Agg.MigrationsCount.Value, err
}

// GetDAppStats -
func (storage *Storage) GetDAppStats(network string, addresses []string, period string) (stats contract.DAppStats, err error) {
	addressMatches := make([]core.Item, len(addresses))
	for i := range addresses {
		addressMatches[i] = core.MatchPhrase("destination", addresses[i])
	}

	matches := []core.Item{
		core.Match("network", network),
		core.Exists("entrypoint"),
		core.Bool(
			core.Should(addressMatches...),
			core.MinimumShouldMatch(1),
		),
		core.Match("status", "applied"),
	}
	r, err := periodToRange(period)
	if err != nil {
		return
	}
	if r != nil {
		matches = append(matches, r)
	}

	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(matches...),
		),
	).Add(
		core.Aggs(
			core.AggItem{Name: "users", Body: core.Cardinality("source.keyword")},
			core.AggItem{Name: "calls", Body: core.Count("indexed_time")},
			core.AggItem{Name: "volume", Body: core.Sum("amount")},
		),
	).Zero()

	var response getDAppStatsResponse
	if err = storage.es.Query([]string{consts.DocOperations}, query, &response); err != nil {
		return
	}

	stats.Calls = int64(response.Aggs.Calls.Value)
	stats.Users = int64(response.Aggs.Users.Value)
	stats.Volume = int64(response.Aggs.Volume.Value)
	return
}

func periodToRange(period string) (core.Item, error) {
	var str string
	switch period {
	case "year":
		str = "now-1y/d"
	case "month":
		str = "now-1M/d"
	case "week":
		str = "now-1w/d"
	case "day":
		str = "now-1d/d"
	case "all":
		return nil, nil
	default:
		return nil, errors.Errorf("Unknown period value: %s", period)
	}
	return core.Item{
		"range": core.Item{
			"timestamp": core.Item{
				"gte": str,
			},
		},
	}, nil
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

// GetByLevels -
func (storage *Storage) GetByLevels(network string, fromLevel, toLevel int64) ([]string, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.Range("level", core.Item{
					"lte": fromLevel,
					"gt":  toLevel,
				}),
			),
		),
	)

	var response core.SearchResponse
	if err := storage.es.Query([]string{consts.DocOperations}, query, &response); err != nil {
		return nil, err
	}

	if response.Hits.Total.Value == 0 {
		return nil, nil
	}

	exists := make(map[string]struct{})
	addresses := make([]string, 0)
	for i := range response.Hits.Hits {
		var op operationAddresses
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &op); err != nil {
			return nil, err
		}
		if _, ok := exists[op.Source]; !ok && helpers.IsContract(op.Source) {
			addresses = append(addresses, op.Source)
			exists[op.Source] = struct{}{}
		}
		if _, ok := exists[op.Destination]; !ok && helpers.IsContract(op.Destination) {
			addresses = append(addresses, op.Destination)
			exists[op.Destination] = struct{}{}
		}
	}

	return addresses, nil
}

// GetProjectsLastContract -
func (storage *Storage) GetProjectsLastContract() ([]contract.Contract, error) {
	query := core.NewQuery().Add(
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
	if err := storage.es.Query([]string{consts.DocContracts}, query, &response); err != nil {
		return nil, err
	}

	if len(response.Agg.Projects.Buckets) == 0 {
		return nil, core.NewRecordNotFoundError(consts.DocContracts, "")
	}

	contracts := make([]contract.Contract, len(response.Agg.Projects.Buckets))
	for i := range response.Agg.Projects.Buckets {
		if err := json.Unmarshal(response.Agg.Projects.Buckets[i].Last.Hits.Hits[0].Source, &contracts[i]); err != nil {
			return nil, err
		}
	}
	return contracts, nil
}

// GetSameContracts -
func (storage *Storage) GetSameContracts(c contract.Contract, size, offset int64) (pcr contract.SameResponse, err error) {
	if c.Fingerprint == nil {
		return pcr, errors.Errorf("Invalid contract data")
	}

	if size == 0 {
		size = consts.DefaultSize
	} else if size+offset > core.MaxQuerySize {
		size = core.MaxQuerySize - offset
	}

	q := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.MatchPhrase("hash", c.Hash),
			),
			core.MustNot(
				core.MatchPhrase("address", c.Address),
			),
		),
	).Sort("last_action", "desc").Size(size).From(offset)

	var response core.SearchResponse
	if err = storage.es.Query([]string{consts.DocContracts}, q, &response); err != nil {
		return
	}

	if len(response.Hits.Hits) == 0 {
		return pcr, core.NewRecordNotFoundError(consts.DocContracts, "")
	}

	contracts := make([]contract.Contract, len(response.Hits.Hits))
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
	if err = storage.es.Query([]string{consts.DocContracts}, query, &response); err != nil {
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
	if err := storage.es.Query([]string{consts.DocContracts}, query, &response); err != nil {
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
	).Sort("timestamp", "desc").Size(size)

	if offset != 0 {
		query = query.From(offset)
	}

	var response core.SearchResponse
	if err := storage.es.Query([]string{consts.DocContracts}, query, &response); err != nil {
		return nil, 0, err
	}

	contracts := make([]contract.Contract, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &contracts[i]); err != nil {
			return nil, 0, err
		}
	}
	return contracts, response.Hits.Total.Value, nil
}
