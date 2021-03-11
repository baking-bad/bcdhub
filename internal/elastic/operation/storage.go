package operation

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd"
	constants "github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
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

func (storage *Storage) getContractOPG(address, network string, size uint64, filters map[string]interface{}) ([]opgForContract, error) {
	if size == 0 || size > core.MaxQuerySize {
		size = consts.DefaultSize
	}

	filtersString, err := prepareOperationFilters(filters)
	if err != nil {
		return nil, err
	}

	sqlString := fmt.Sprintf(`SELECT hash, counter
		FROM operation 
		WHERE (source = '%s' OR destination = '%s') AND network = '%s' %s 
		GROUP BY hash, counter, level
		ORDER BY level DESC
		LIMIT %d`, address, address, network, filtersString, size)

	var response core.SQLResponse
	if err := storage.es.ExecuteSQL(sqlString, &response); err != nil {
		return nil, err
	}

	resp := make([]opgForContract, 0)
	for i := range response.Rows {
		resp = append(resp, opgForContract{
			hash:    response.Rows[i][0].(string),
			counter: int64(response.Rows[i][1].(float64)),
		})
	}

	return resp, nil
}

func prepareOperationFilters(filters map[string]interface{}) (s string, err error) {
	for k, v := range filters {
		if v != "" {
			s += " AND "
			switch k {
			case "from":
				s += fmt.Sprintf("timestamp >= %d", v)
			case "to":
				s += fmt.Sprintf("timestamp <= %d", v)
			case "entrypoints":
				s += fmt.Sprintf("entrypoint IN (%s)", v)
			case "last_id":
				s += fmt.Sprintf("indexed_time < %s", v)
			case "status":
				s += fmt.Sprintf("status IN (%s)", v)
			default:
				return "", errors.Errorf("Unknown operation filter: %s %v", k, v)
			}
		}
	}
	return
}

// GetByContract -
func (storage *Storage) GetByContract(network, address string, size uint64, filters map[string]interface{}) (po operation.Pageable, err error) {
	opg, err := storage.getContractOPG(address, network, size, filters)
	if err != nil {
		return
	}

	s := make([]core.Item, len(opg))
	for i := range opg {
		s[i] = core.Bool(core.Filter(
			core.Match("hash", opg[i].hash),
			core.Term("counter", opg[i].counter),
		))
	}
	b := core.Bool(
		core.Should(s...),
		core.Filter(
			core.Match("network", network),
		),
		core.MinimumShouldMatch(1),
	)
	query := core.NewQuery().
		Query(b).
		Add(
			core.Aggs(core.AggItem{Name: "last_id", Body: core.Min("indexed_time")}),
		).
		Add(core.Item{
			"sort": core.Item{
				"_script": core.Item{
					"type": "number",
					"script": core.Item{
						"lang":   "painless",
						"source": "doc['level'].value * 10000000000L + (doc['counter'].value) * 1000L + (doc['internal'].value ? (998L - doc['nonce'].value) : 999L)",
					},
					"order": "desc",
				},
			},
		}).All()

	var response getByContract
	if err = storage.es.Query([]string{models.DocOperations}, query, &response); err != nil {
		return
	}

	ops := make([]operation.Operation, len(response.Hist.Hits))
	for i := range response.Hist.Hits {
		if err = json.Unmarshal(response.Hist.Hits[i].Source, &ops[i]); err != nil {
			return
		}
		ops[i].ID = response.Hist.Hits[i].ID
	}

	po.Operations = ops
	po.LastID = fmt.Sprintf("%.0f", response.Agg.LastID.Value)
	return
}

// Last -
func (storage *Storage) Last(network, address string, indexedTime int64) (op operation.Operation, err error) {
	query := core.NewQuery().
		Query(
			core.Bool(
				core.Filter(
					core.MatchPhrase("destination", address),
					core.Range("indexed_time", core.Item{"lt": indexedTime}),
					core.Term("network", network),
					core.Term("status", "applied"),
				),
				core.MustNot(
					core.Term("deffated_storage", ""),
				),
			),
		).Sort("indexed_time", "desc").One()

	var response core.SearchResponse
	if err = storage.es.Query([]string{models.DocOperations}, query, &response); err != nil {
		return
	}

	if response.Hits.Total.Value == 0 {
		return op, core.NewRecordNotFoundError(models.DocOperations, "")
	}
	err = json.Unmarshal(response.Hits.Hits[0].Source, &op)
	op.ID = response.Hits.Hits[0].ID
	return
}

// Get -
func (storage *Storage) Get(filters map[string]interface{}, size int64, sort bool) ([]operation.Operation, error) {
	operations := make([]operation.Operation, 0)

	query := core.FiltersToQuery(filters)

	if sort {
		query.Add(core.Item{
			"sort": core.Item{
				"_script": core.Item{
					"type": "number",
					"script": core.Item{
						"lang":   "painless",
						"source": "doc['level'].value * 10000000000L + (doc['counter'].value) * 1000L + (doc['internal'].value ? (998L - doc['nonce'].value) : 999L)",
					},
					"order": "desc",
				},
			},
		})
	}

	scrollSize := size
	if consts.DefaultScrollSize < scrollSize || scrollSize == 0 {
		scrollSize = consts.DefaultScrollSize
	}

	ctx := core.NewScrollContext(storage.es, query, size, scrollSize)
	err := ctx.Get(&operations)
	return operations, err
}

// GetStats -
func (storage *Storage) GetStats(network, address string) (stats operation.Stats, err error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.Bool(
					core.Should(
						core.MatchPhrase("source", address),
						core.MatchPhrase("destination", address),
					),
					core.MinimumShouldMatch(1),
				),
			),
		),
	).Add(
		core.Aggs(
			core.AggItem{
				Name: "opg", Body: core.Count("hash.keyword"),
			},
			core.AggItem{
				Name: "last_action", Body: core.Max("timestamp"),
			},
		),
	).Zero()

	var response getOperationsStatsResponse
	if err = storage.es.Query([]string{models.DocOperations}, query, &response); err != nil {
		return
	}

	stats.Count = response.Aggs.OPG.Value
	stats.LastAction = response.Aggs.LastAction.Value
	return
}

// GetContract24HoursVolume -
func (storage *Storage) GetContract24HoursVolume(network, address string, entrypoints []string) (float64, error) {
	filter := []core.Item{
		core.MatchPhrase("destination", address),
		core.Term("network", network),
		core.Term("status", constants.Applied),
		core.Range("timestamp", core.Item{
			"lte": "now",
			"gt":  "now-24h",
		}),
	}

	if len(entrypoints) > 0 {
		filter = append(filter, core.In("entrypoint.keyword", entrypoints))
	}

	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(filter...),
		),
	).Add(
		core.Aggs(
			core.AggItem{Name: "volume", Body: core.Sum("amount")},
		),
	).Zero()

	var response aggVolumeSumResponse
	if err := storage.es.Query([]string{models.DocOperations}, query, &response); err != nil {
		return 0, err
	}

	return response.Aggs.Result.Value, nil
}

// GetTokensStats -
func (storage *Storage) GetTokensStats(network string, addresses, entrypoints []string) (map[string]operation.TokenUsageStats, error) {
	addressFilters := make([]core.Item, len(addresses))
	for i := range addresses {
		addressFilters[i] = core.MatchPhrase("destination", addresses[i])
	}

	entrypointFilters := make([]core.Item, len(entrypoints))
	for i := range entrypoints {
		entrypointFilters[i] = core.MatchPhrase("entrypoint", entrypoints[i])
	}

	query := core.NewQuery().Query(
		core.Bool(
			core.Must(
				core.Match("network", network),
				core.Bool(
					core.Should(addressFilters...),
					core.MinimumShouldMatch(1),
				),
				core.Bool(
					core.Should(entrypointFilters...),
					core.MinimumShouldMatch(1),
				),
			),
		),
	).Add(
		core.Aggs(
			core.AggItem{
				Name: "body",
				Body: core.Composite(
					core.MaxQuerySize,
					core.AggItem{
						Name: "destination", Body: core.TermsAgg("destination.keyword", 0),
					},
					core.AggItem{
						Name: "entrypoint", Body: core.TermsAgg("entrypoint.keyword", 0),
					},
				).Extend(
					core.Aggs(
						core.AggItem{
							Name: "average_consumed_gas", Body: core.Avg("result.consumed_gas"),
						},
					),
				),
			},
		),
	).Zero()

	var response getTokensStatsResponse
	if err := storage.es.Query([]string{models.DocOperations}, query, &response); err != nil {
		return nil, err
	}

	usageStats := make(map[string]operation.TokenUsageStats)
	for _, bucket := range response.Aggs.Body.Buckets {
		usage := operation.TokenMethodUsageStats{
			Count:       bucket.DocCount,
			ConsumedGas: int64(bucket.AVG.Value),
		}

		if _, ok := usageStats[bucket.Key.Destination]; !ok {
			usageStats[bucket.Key.Destination] = make(operation.TokenUsageStats)
		}
		usageStats[bucket.Key.Destination][bucket.Key.Entrypoint] = usage
	}

	return usageStats, nil
}

// GetParticipatingContracts -
func (storage *Storage) GetParticipatingContracts(network string, fromLevel, toLevel int64) ([]string, error) {
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
	if err := storage.es.Query([]string{models.DocOperations}, query, &response); err != nil {
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
		if _, ok := exists[op.Source]; !ok && bcd.IsContract(op.Source) {
			addresses = append(addresses, op.Source)
			exists[op.Source] = struct{}{}
		}
		if _, ok := exists[op.Destination]; !ok && bcd.IsContract(op.Destination) {
			addresses = append(addresses, op.Destination)
			exists[op.Destination] = struct{}{}
		}
	}

	return addresses, nil
}

// RecalcStats -
func (storage *Storage) RecalcStats(network, address string) (stats operation.ContractStats, err error) {
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
			},
		},
	).Zero()
	var response recalcContractStatsResponse
	if err = storage.es.Query([]string{models.DocOperations}, query, &response); err != nil {
		return
	}

	stats.LastAction = time.Unix(0, response.Aggs.LastAction.Value*1000000).UTC()
	stats.Balance = response.Aggs.Balance.Value
	stats.TxCount = response.Aggs.TxCount.Value
	return
}

// GetDAppStats -
func (storage *Storage) GetDAppStats(network string, addresses []string, period string) (stats operation.DAppStats, err error) {
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
	if err = storage.es.Query([]string{models.DocOperations}, query, &response); err != nil {
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
