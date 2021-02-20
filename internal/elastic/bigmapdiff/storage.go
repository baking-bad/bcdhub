package bigmapdiff

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
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

// CurrentByKey -
func (storage *Storage) CurrentByKey(network, keyHash string, ptr int64) (data bigmapdiff.BigMapDiff, err error) {
	if ptr < 0 {
		err = errors.Errorf("Invalid pointer value: %d", ptr)
		return
	}
	mustQuery := core.Must(
		core.MatchPhrase("network", network),
		core.MatchPhrase("key_hash", keyHash),
		core.Term("ptr", ptr),
	)
	b := core.Bool(mustQuery)

	query := core.NewQuery().Query(b).Sort("level", "desc").One()

	var response core.SearchResponse
	if err = storage.es.Query([]string{models.DocBigMapDiff}, query, &response); err != nil {
		return
	}

	if response.Hits.Total.Value == 0 {
		return data, core.NewRecordNotFoundError(models.DocBigMapDiff, "")
	}
	err = json.Unmarshal(response.Hits.Hits[0].Source, &data)
	return
}

// GetForAddress -
func (storage *Storage) GetForAddress(address string) ([]bigmapdiff.BigMapDiff, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Must(
				core.MatchPhrase("address", address),
			),
		),
	).Add(
		core.Aggs(
			core.AggItem{
				Name: "keys",
				Body: core.Item{
					"terms": core.Item{
						"field": "key_hash.keyword",
						"size":  core.MaxQuerySize, // TODO: arbitrary number of keys
					},
					"aggs": core.Item{
						"top_key": core.TopHits(1, "indexed_time", "desc"),
					},
				},
			},
		),
	).Zero()

	var response getBigMapDiffsWithKeysResponse
	if err := storage.es.Query([]string{models.DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}
	arr := response.Agg.Keys.Buckets
	diffs := make([]bigmapdiff.BigMapDiff, len(arr))
	for i := range arr {
		if err := json.Unmarshal(arr[i].TopKey.Hits.Hits[0].Source, &diffs[i]); err != nil {
			return nil, err
		}
		diffs[i].ID = arr[i].TopKey.Hits.Hits[0].ID
	}
	return diffs, nil
}

// GetByAddress -
func (storage *Storage) GetByAddress(network, address string) (response []bigmapdiff.BigMapDiff, err error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.MatchPhrase("address", address),
			),
		),
	).Sort("indexed_time", "desc")

	err = storage.es.GetAllByQuery(query, &response)
	return
}

// GetValuesByKey -
func (storage *Storage) GetValuesByKey(keyHash string) ([]bigmapdiff.BigMapDiff, error) {
	mustQuery := core.Must(
		core.MatchPhrase("key_hash", keyHash),
	)
	b := core.Bool(mustQuery)

	query := core.NewQuery().Query(b).Add(
		core.Aggs(
			core.AggItem{
				Name: "keys",
				Body: core.Item{
					"terms": core.Item{
						"script": core.Item{
							"source": "doc['network.keyword'].value + doc['address.keyword'].value + String.format('%d', new def[] {doc['ptr'].value})",
						},
					},
					"aggs": core.Item{
						"top_key": core.TopHits(1, "indexed_time", "desc"),
					},
				},
			},
		),
	).Zero()

	var response getBigMapDiffsWithKeysResponse
	if err := storage.es.Query([]string{models.DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}

	bmd := make([]bigmapdiff.BigMapDiff, len(response.Agg.Keys.Buckets))
	for i, item := range response.Agg.Keys.Buckets {
		if err := json.Unmarshal(item.TopKey.Hits.Hits[0].Source, &bmd[i]); err != nil {
			return nil, err
		}
	}
	return bmd, nil
}

// Count -
func (storage *Storage) Count(network string, ptr int64) (int64, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.Term("ptr", ptr),
			),
		),
	).Add(
		core.Aggs(core.AggItem{
			Name: "count",
			Body: core.Cardinality("key_hash.keyword"),
		}),
	).Zero()

	var response getBigMapDiffsCountResponse
	if err := storage.es.Query([]string{models.DocBigMapDiff}, query, &response); err != nil {
		return 0, err
	}
	return response.Agg.Count.Value, nil
}

// Previous -
func (storage *Storage) Previous(filters []bigmapdiff.BigMapDiff, indexedTime int64, address string) ([]bigmapdiff.BigMapDiff, error) {
	shouldData := make([]core.Item, len(filters))
	for i := range filters {
		shouldData[i] = core.Bool(core.Filter(
			core.MatchPhrase("key_hash", filters[i].KeyHash),
		))
	}
	b := core.Bool(
		core.Should(shouldData...),
		core.Filter(
			core.MatchPhrase("address", address),
			core.Range("indexed_time", core.Item{"lt": indexedTime}),
		),
		core.MinimumShouldMatch(1),
	)

	query := core.NewQuery().Query(b).
		Add(
			core.Aggs(
				core.AggItem{
					Name: "keys",
					Body: core.Item{
						"terms": core.Item{
							"field": "key_hash.keyword",
							"size":  core.MaxQuerySize,
						},
						"aggs": core.Item{
							"top_key": core.TopHits(1, "indexed_time", "desc"),
						},
					},
				},
			),
		).
		Sort("indexed_time", "desc").Zero()

	var response getBigMapDiffsWithKeysResponse
	if err := storage.es.Query([]string{models.DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}

	arr := response.Agg.Keys.Buckets
	diffs := make([]bigmapdiff.BigMapDiff, 0)
	for i := range arr {
		var b bigmapdiff.BigMapDiff
		if err := json.Unmarshal(arr[i].TopKey.Hits.Hits[0].Source, &b); err != nil {
			return nil, err
		}
		if b.Value != nil {
			b.ID = arr[i].TopKey.Hits.Hits[0].ID
			diffs = append(diffs, b)
		}
	}
	return diffs, nil
}

// GetUniqueByOperationID -
func (storage *Storage) GetUniqueByOperationID(operationID string) ([]bigmapdiff.BigMapDiff, error) {
	query := core.NewQuery().
		Query(
			core.Bool(
				core.Filter(
					core.MatchPhrase("operation_id", operationID),
				),
			),
		).
		Add(
			core.Aggs(
				core.AggItem{
					Name: "keys",
					Body: core.Composite(
						core.MaxQuerySize,
						core.AggItem{
							Name: "ptr",
							Body: core.TermsAgg("ptr", 0),
						},
						core.AggItem{
							Name: "key_hash",
							Body: core.TermsAgg("key_hash.keyword", 0),
						},
					).Extend(
						core.Aggs(
							core.AggItem{
								Name: "top_key",
								Body: core.TopHits(1, "indexed_time", "desc"),
							},
						),
					),
				},
			),
		).Zero()

	var response getBigMapDiffsWithKeysResponse
	if err := storage.es.Query([]string{models.DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}
	arr := response.Agg.Keys.Buckets
	diffs := make([]bigmapdiff.BigMapDiff, len(arr))
	for i := range arr {
		if err := json.Unmarshal(arr[i].TopKey.Hits.Hits[0].Source, &diffs[i]); err != nil {
			return nil, err
		}
		diffs[i].ID = arr[i].TopKey.Hits.Hits[0].ID
	}
	return diffs, nil
}

// GetByPtrAndKeyHash -
func (storage *Storage) GetByPtrAndKeyHash(ptr int64, network, keyHash string, size, offset int64) ([]bigmapdiff.BigMapDiff, int64, error) {
	if ptr < 0 {
		return nil, 0, errors.Errorf("Invalid pointer value: %d", ptr)
	}
	mustQuery := core.Must(
		core.MatchPhrase("network", network),
		core.MatchPhrase("key_hash", keyHash),
		core.Term("ptr", ptr),
	)
	b := core.Bool(mustQuery)

	if size == 0 {
		size = consts.DefaultSize
	}

	query := core.NewQuery().Query(b).Sort("level", "desc").Size(size).From(offset)

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocBigMapDiff}, query, &response); err != nil {
		return nil, 0, err
	}

	result := make([]bigmapdiff.BigMapDiff, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &result[i]); err != nil {
			return nil, 0, err
		}
		result[i].ID = response.Hits.Hits[i].ID
	}

	return result, response.Hits.Total.Value, nil
}

// GetByOperationID -
func (storage *Storage) GetByOperationID(operationID string) ([]*bigmapdiff.BigMapDiff, error) {
	query := core.NewQuery().
		Query(
			core.Bool(
				core.Must(
					core.MatchPhrase("operation_id", operationID),
				),
			),
		).All()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}
	result := make([]*bigmapdiff.BigMapDiff, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &result[i]); err != nil {
			return nil, err
		}
		result[i].ID = response.Hits.Hits[i].ID
	}
	return result, nil
}

// GetByPtr -
func (storage *Storage) GetByPtr(address, network string, ptr int64) ([]bigmapdiff.BigMapDiff, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.MatchPhrase("address", address),
				core.Term("ptr", ptr),
			),
		),
	).Add(
		core.Aggs(core.AggItem{
			Name: "keys",
			Body: core.Item{
				"terms": core.Item{
					"field": "key_hash.keyword",
					"size":  core.MaxQuerySize,
				},
				"aggs": core.Item{
					"top_key": core.TopHits(1, "indexed_time", "desc"),
				},
			},
		}),
	).Sort("indexed_time", "desc").Zero()

	var response getBigMapDiffsWithKeysResponse
	if err := storage.es.Query([]string{models.DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}
	bmd := make([]bigmapdiff.BigMapDiff, len(response.Agg.Keys.Buckets))
	for i := range response.Agg.Keys.Buckets {
		if err := json.Unmarshal(response.Agg.Keys.Buckets[i].TopKey.Hits.Hits[0].Source, &bmd[i]); err != nil {
			return nil, err
		}
	}
	return bmd, nil
}

// Get -
func (storage *Storage) Get(ctx bigmapdiff.GetContext) ([]bigmapdiff.Bucket, error) {
	if *ctx.Ptr < 0 {
		return nil, errors.Errorf("Invalid pointer value: %d", *ctx.Ptr)
	}

	query := buildGetContext(&ctx)

	var response getBigMapDiffsWithKeysResponse
	if err := storage.es.Query([]string{models.DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}

	arr := response.Agg.Keys.Buckets
	if int64(len(arr)) < ctx.Offset {
		return nil, nil
	}

	if int64(len(arr)) < ctx.To {
		ctx.To = int64(len(arr))
	}

	arr = arr[ctx.Offset:ctx.To]
	result := make([]bigmapdiff.Bucket, len(arr))
	for i := range arr {
		if err := json.Unmarshal(arr[i].TopKey.Hits.Hits[0].Source, &result[i]); err != nil {
			return nil, err
		}
		result[i].ID = arr[i].TopKey.Hits.Hits[0].ID
		result[i].Count = arr[i].DocCount
	}
	return result, nil
}
