package elastic

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

type getBigMapDiffsWithKeysResponse struct {
	Agg struct {
		Keys struct {
			Buckets []struct {
				DocCount int64 `json:"doc_count"`
				TopKey   struct {
					Hits HitsArray `json:"hits"`
				} `json:"top_key"`
			} `json:"buckets"`
		} `json:"keys"`
	} `json:"aggregations"`
}

// GetBigMapDiffsUniqueByOperationID -
func (e *Elastic) GetBigMapDiffsUniqueByOperationID(operationID string) ([]models.BigMapDiff, error) {
	query := newQuery().
		Query(
			boolQ(
				filter(
					matchPhrase("operation_id", operationID),
				),
			),
		).
		Add(
			aggs(
				aggItem{
					"keys",
					composite(
						maxQuerySize,
						aggItem{
							"ptr", termsAgg("ptr", 0),
						},
						aggItem{
							"key_hash", termsAgg("key_hash.keyword", 0),
						},
					).Extend(
						aggs(
							aggItem{
								"top_key", topHits(1, "indexed_time", "desc"),
							},
						),
					),
				},
			),
		).Zero()

	var response getBigMapDiffsWithKeysResponse
	if err := e.query([]string{DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}
	arr := response.Agg.Keys.Buckets
	diffs := make([]models.BigMapDiff, len(arr))
	for i := range arr {
		if err := json.Unmarshal(arr[i].TopKey.Hits.Hits[0].Source, &diffs[i]); err != nil {
			return nil, err
		}
		diffs[i].ID = arr[i].TopKey.Hits.Hits[0].ID
	}
	return diffs, nil
}

// GetBigMapDiffsPrevious -
func (e *Elastic) GetBigMapDiffsPrevious(filters []models.BigMapDiff, indexedTime int64, address string) ([]models.BigMapDiff, error) {
	shouldData := make([]qItem, len(filters))
	for i := range filters {
		shouldData[i] = boolQ(filter(
			matchPhrase("key_hash", filters[i].KeyHash),
			matchPhrase("bin_path", filters[i].BinPath),
		))
	}
	b := boolQ(
		should(shouldData...),
		filter(
			matchPhrase("address", address),
			rangeQ("indexed_time", qItem{"lt": indexedTime}),
		),
		minimumShouldMatch(1),
	)

	query := newQuery().Query(b).
		Add(
			aggs(
				aggItem{
					"keys", qItem{
						"terms": qItem{
							"field": "key_hash.keyword",
							"size":  maxQuerySize,
						},
						"aggs": qItem{
							"top_key": topHits(1, "indexed_time", "desc"),
						},
					},
				},
			),
		).
		Sort("indexed_time", "desc").Zero()

	var response getBigMapDiffsWithKeysResponse
	if err := e.query([]string{DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}

	arr := response.Agg.Keys.Buckets
	diffs := make([]models.BigMapDiff, 0)
	for i := range arr {
		var b models.BigMapDiff
		if err := json.Unmarshal(arr[i].TopKey.Hits.Hits[0].Source, &b); err != nil {
			return nil, err
		}
		if b.Value != "" {
			b.ID = arr[i].TopKey.Hits.Hits[0].ID
			diffs = append(diffs, b)
		}
	}
	return diffs, nil
}

// GetBigMapDiffsForAddress -
func (e *Elastic) GetBigMapDiffsForAddress(address string) ([]models.BigMapDiff, error) {
	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("address", address),
			),
		),
	).Add(
		aggs(
			aggItem{
				"keys", qItem{
					"terms": qItem{
						"field": "key_hash.keyword",
						"size":  maxQuerySize, // TODO: arbitrary number of keys
					},
					"aggs": qItem{
						"top_key": topHits(1, "indexed_time", "desc"),
					},
				},
			},
		),
	).Zero()

	var response getBigMapDiffsWithKeysResponse
	if err := e.query([]string{DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}
	arr := response.Agg.Keys.Buckets
	diffs := make([]models.BigMapDiff, len(arr))
	for i := range arr {
		if err := json.Unmarshal(arr[i].TopKey.Hits.Hits[0].Source, &diffs[i]); err != nil {
			return nil, err
		}
		diffs[i].ID = arr[i].TopKey.Hits.Hits[0].ID
	}
	return diffs, nil
}

// GetBigMapKeysContext -
type GetBigMapKeysContext struct {
	Network string
	Ptr     *int64
	Query   string
	Size    int64
	Offset  int64
	Level   *int64

	to int64
}

func (ctx *GetBigMapKeysContext) build() base {
	filters := make([]qItem, 0)

	if ctx.Ptr != nil {
		filters = append(filters, term("ptr", *ctx.Ptr))
	}
	if ctx.Network != "" {
		filters = append(filters, matchQ("network", ctx.Network))
	}

	if ctx.Query != "" {
		filters = append(filters, queryString(fmt.Sprintf("*%s*", ctx.Query), []string{"key", "key_hash", "key_strings", "bin_path"}))
	}

	if ctx.Size == 0 {
		ctx.Size = defaultSize
	}

	if ctx.Level != nil {
		filters = append(filters, NewLessThanEqRange(*ctx.Level).build())
	}

	ctx.to = ctx.Size + ctx.Offset
	b := boolQ(
		must(filters...),
	)
	return newQuery().Query(b).Add(
		aggs(aggItem{
			"keys", qItem{
				"terms": qItem{
					"field": "key_hash.keyword",
					"size":  ctx.to,
					"order": qItem{
						"bucketsSort": "desc",
					},
				},
				"aggs": qItem{
					"top_key":     topHits(1, "indexed_time", "desc"),
					"bucketsSort": max("indexed_time"),
				},
			},
		}),
	).Sort("indexed_time", "desc").Zero()
}

// GetBigMapKeys -
func (e *Elastic) GetBigMapKeys(ctx GetBigMapKeysContext) ([]BigMapDiff, error) {
	if *ctx.Ptr < 0 {
		return nil, errors.Errorf("Invalid pointer value: %d", *ctx.Ptr)
	}

	var response getBigMapDiffsWithKeysResponse
	if err := e.query([]string{DocBigMapDiff}, ctx.build(), &response); err != nil {
		return nil, err
	}

	arr := response.Agg.Keys.Buckets
	if int64(len(arr)) < ctx.Offset {
		return nil, nil
	}

	if int64(len(arr)) < ctx.to {
		ctx.to = int64(len(arr))
	}

	arr = arr[ctx.Offset:ctx.to]
	result := make([]BigMapDiff, len(arr))
	for i := range arr {
		var b models.BigMapDiff
		if err := json.Unmarshal(arr[i].TopKey.Hits.Hits[0].Source, &b); err != nil {
			return nil, err
		}
		b.ID = arr[i].TopKey.Hits.Hits[0].ID
		if err := result[i].FromModel(&b); err != nil {
			return nil, err
		}
		result[i].Count = arr[i].DocCount
	}
	return result, nil
}

// GetBigMapDiffsByPtrAndKeyHash -
func (e *Elastic) GetBigMapDiffsByPtrAndKeyHash(ptr int64, network, keyHash string, size, offset int64) ([]BigMapDiff, int64, error) {
	if ptr < 0 {
		return nil, 0, errors.Errorf("Invalid pointer value: %d", ptr)
	}
	mustQuery := must(
		matchPhrase("network", network),
		matchPhrase("key_hash", keyHash),
		term("ptr", ptr),
	)
	b := boolQ(mustQuery)

	if size == 0 {
		size = defaultSize
	}

	query := newQuery().Query(b).Sort("level", "desc").Size(size).From(offset)

	var response SearchResponse
	if err := e.query([]string{DocBigMapDiff}, query, &response); err != nil {
		return nil, 0, err
	}

	result := make([]BigMapDiff, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		var b models.BigMapDiff
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &b); err != nil {
			return nil, 0, err
		}
		b.ID = response.Hits.Hits[i].ID
		if err := result[i].FromModel(&b); err != nil {
			return nil, 0, err
		}
	}

	return result, response.Hits.Total.Value, nil
}

// GetBigMapDiffsByOperationID -
func (e *Elastic) GetBigMapDiffsByOperationID(operationID string) ([]*models.BigMapDiff, error) {
	query := newQuery().
		Query(
			boolQ(
				must(
					matchPhrase("operation_id", operationID),
				),
			),
		).All()

	var response SearchResponse
	if err := e.query([]string{DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}
	result := make([]*models.BigMapDiff, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &result[i]); err != nil {
			return nil, err
		}
		result[i].ID = response.Hits.Hits[i].ID
	}
	return result, nil
}

// GetBigMapDiffsByPtr -
func (e *Elastic) GetBigMapDiffsByPtr(address, network string, ptr int64) ([]models.BigMapDiff, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				matchPhrase("address", address),
				term("ptr", ptr),
			),
		),
	).Add(
		aggs(aggItem{
			"keys", qItem{
				"terms": qItem{
					"field": "key_hash.keyword",
					"size":  maxQuerySize,
				},
				"aggs": qItem{
					"top_key": topHits(1, "indexed_time", "desc"),
				},
			},
		}),
	).Sort("indexed_time", "desc").Zero()

	var response getBigMapDiffsWithKeysResponse
	if err := e.query([]string{DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}
	bmd := make([]models.BigMapDiff, len(response.Agg.Keys.Buckets))
	for i := range response.Agg.Keys.Buckets {
		if err := json.Unmarshal(response.Agg.Keys.Buckets[i].TopKey.Hits.Hits[0].Source, &bmd[i]); err != nil {
			return nil, err
		}
	}
	return bmd, nil
}

// GetBigMapsForAddress -
func (e *Elastic) GetBigMapsForAddress(network, address string) (response []models.BigMapDiff, err error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				matchPhrase("address", address),
			),
		),
	).Sort("indexed_time", "desc")

	err = e.getAllByQuery(query, &response)
	return
}

// GetBigMapHistory -
func (e *Elastic) GetBigMapHistory(ptr int64, network string) (response []models.BigMapAction, err error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
			should(
				term("source_ptr", ptr),
				term("destination_ptr", ptr),
			),
			minimumShouldMatch(1),
		),
	).Sort("indexed_time", "desc")

	err = e.getAllByQuery(query, &response)
	return
}

// GetBigMapKey -
func (e *Elastic) GetBigMapKey(network, keyHash string, ptr int64) (data BigMapDiff, err error) {
	if ptr < 0 {
		err = errors.Errorf("Invalid pointer value: %d", ptr)
		return
	}
	mustQuery := must(
		matchPhrase("network", network),
		matchPhrase("key_hash", keyHash),
		term("ptr", ptr),
	)
	b := boolQ(mustQuery)

	query := newQuery().Query(b).Sort("level", "desc").One()

	var response SearchResponse
	if err = e.query([]string{DocBigMapDiff}, query, &response); err != nil {
		return
	}

	if response.Hits.Total.Value == 0 {
		return data, NewRecordNotFoundError(DocBigMapDiff, "", query)
	}
	err = json.Unmarshal(response.Hits.Hits[0].Source, &data)
	return
}

// GetBigMapValuesByKey -
func (e *Elastic) GetBigMapValuesByKey(keyHash string) ([]BigMapDiff, error) {
	mustQuery := must(
		matchPhrase("key_hash", keyHash),
	)
	b := boolQ(mustQuery)

	query := newQuery().Query(b).Add(
		aggs(
			aggItem{
				"keys", qItem{
					"terms": qItem{
						"script": qItem{
							"source": "doc['network.keyword'].value + doc['address.keyword'].value + String.format('%d', new def[] {doc['ptr'].value})",
						},
					},
					"aggs": qItem{
						"top_key": topHits(1, "indexed_time", "desc"),
					},
				},
			},
		),
	).Zero()

	var response getBigMapDiffsWithKeysResponse
	if err := e.query([]string{DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}

	bmd := make([]BigMapDiff, len(response.Agg.Keys.Buckets))
	for i, item := range response.Agg.Keys.Buckets {
		if err := json.Unmarshal(item.TopKey.Hits.Hits[0].Source, &bmd[i]); err != nil {
			return nil, err
		}
	}
	return bmd, nil
}

type getBigMapDiffsCountResponse struct {
	Agg struct {
		Count intValue `json:"count"`
	} `json:"aggregations"`
}

// GetBigMapDiffsCount -
func (e *Elastic) GetBigMapDiffsCount(network string, ptr int64) (int64, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				term("ptr", ptr),
			),
		),
	).Add(
		aggs(aggItem{
			"count", cardinality("key_hash.keyword"),
		}),
	).Zero()

	var response getBigMapDiffsCountResponse
	if err := e.query([]string{DocBigMapDiff}, query, &response); err != nil {
		return 0, err
	}
	return response.Agg.Count.Value, nil
}
