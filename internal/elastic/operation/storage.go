package operation

import (
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
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

type opgForContract struct {
	hash    string
	counter int64
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
	if err = storage.es.Query([]string{consts.DocOperations}, query, &response); err != nil {
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
				core.Must(
					core.MatchPhrase("destination", address),
					core.MatchPhrase("network", network),
				),
				core.Filter(
					core.Range("indexed_time", core.Item{"lt": indexedTime}),
					core.Term("status", "applied"),
				),
				core.MustNot(
					core.Term("deffated_storage", ""),
				),
			),
		).Sort("indexed_time", "desc").One()

	var response core.SearchResponse
	if err = storage.es.Query([]string{consts.DocOperations}, query, &response); err != nil {
		return
	}

	if response.Hits.Total.Value == 0 {
		return op, core.NewRecordNotFoundError(consts.DocOperations, "")
	}
	err = json.Unmarshal(response.Hits.Hits[0].Source, &op)
	op.ID = response.Hits.Hits[0].ID
	return
}

type operationAddresses struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
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
	if err = storage.es.Query([]string{consts.DocOperations}, query, &response); err != nil {
		return
	}

	stats.Count = response.Aggs.OPG.Value
	stats.LastAction = response.Aggs.LastAction.Value
	return
}

// GetContract24HoursVolume -
func (storage *Storage) GetContract24HoursVolume(network, address string, entrypoints []string) (float64, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				boolQ(
					should(
						matchPhrase("destination", address),
						matchPhrase("source", address),
					),
					minimumShouldMatch(1),
				),
				term("network", network),
				term("status", consts.Applied),
				rangeQ("timestamp", qItem{
					"lte": "now",
					"gt":  "now-24h",
				}),
				in("entrypoint.keyword", entrypoints),
			),
		),
	).Add(
		aggs(
			aggItem{"volume", sum("amount")},
		),
	).Zero()

	var response aggVolumeSumResponse
	if err := e.query([]string{consts.DocOperations}, query, &response); err != nil {
		return 0, err
	}

	return response.Aggs.Result.Value, nil
}
