package elastic

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// GetBigMapDiffsByOperationID -
func (e *Elastic) GetBigMapDiffsByOperationID(operationID string) ([]models.BigMapDiff, error) {
	query := newQuery().
		Query(
			boolQ(
				must(
					matchPhrase("operation_id", operationID),
				),
			),
		).All()

	res, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return nil, err
	}
	response := make([]models.BigMapDiff, 0)
	for _, item := range res.Get("hits.hits").Array() {
		var bmd models.BigMapDiff
		bmd.ParseElasticJSON(item)
		response = append(response, bmd)
	}
	return response, nil
}

// GetBigMapDiffsByKeyHashAndPtr -
func (e *Elastic) GetBigMapDiffsByKeyHashAndPtr(keys []string, ptr []int64, indexedTime int64, address string) ([]models.BigMapDiff, error) {
	shouldData := make([]qItem, len(keys))
	for i := range keys {
		shouldData[i] = boolQ(must(
			matchPhrase("key_hash", keys[i]),
			term("ptr", ptr[i]),
		))
	}
	b := boolQ(
		should(shouldData...),
		must(matchPhrase("address", address)),
		filter(
			rangeQ("indexed_time", qItem{"lt": indexedTime}),
		),
		minimumShouldMatch(1),
	)

	query := newQuery().Query(b).Sort("indexed_time", "desc").All()

	res, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return nil, err
	}

	response := make([]models.BigMapDiff, 0)
	for i := range keys {
		for _, item := range res.Get("hits.hits").Array() {
			keyHash := item.Get("_source.key_hash").String()
			pointer := item.Get("_source.ptr").Int()
			if pointer == ptr[i] && keyHash == keys[i] {
				var bmd models.BigMapDiff
				bmd.ParseElasticJSON(item)
				response = append(response, bmd)
				break
			}
		}
	}
	return response, nil
}

// GetBigMapDiffsForAddress -
func (e *Elastic) GetBigMapDiffsForAddress(address string) ([]models.BigMapDiff, error) {
	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("address", address),
			),
		),
	).Add(aggs(
		"keys", qItem{
			"terms": qItem{
				"field": "key_hash.keyword",
				"size":  maxQuerySize, // TODO: arbitrary number of keys
			},
			"aggs": qItem{
				"top_key": topHits(1, "indexed_time", "desc"),
			},
		},
	)).Zero()

	res, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return nil, err
	}
	response := make([]models.BigMapDiff, 0)
	for _, item := range res.Get("aggregations.keys.buckets").Array() {
		bmd := item.Get("top_key.hits.hits.0")
		var b models.BigMapDiff
		b.ParseElasticJSON(bmd)
		response = append(response, b)
	}
	return response, nil
}

// GetBigMap -
func (e *Elastic) GetBigMap(address string, ptr int64, searchText string, size, offset int64) ([]BigMapDiff, error) {
	mustQuery := []qItem{
		matchPhrase("address", address),
	}
	if searchText != "" {
		mustQuery = append(mustQuery, queryString(searchText, []string{"key", "key_hash", "key_strings"}))
	}

	if ptr != 0 {
		mustQuery = append(mustQuery, term("ptr", ptr))
	}

	b := boolQ(must(mustQuery...))

	if ptr == 0 {
		b.Get("bool").Extend(
			notMust(
				qItem{
					"exists": qItem{
						"field": "ptr",
					},
				},
			),
		)
	}

	if size == 0 {
		size = defaultSize
	}

	to := size + offset
	query := newQuery().Query(b).Add(
		aggs("keys", qItem{
			"terms": qItem{
				"field": "key_hash.keyword",
				"size":  to,
				"order": qItem{
					"bucketsSort": "desc",
				},
			},
			"aggs": qItem{
				"top_key":     topHits(1, "indexed_time", "desc"),
				"bucketsSort": max("indexed_time"),
			},
		}),
	).Sort("indexed_time", "desc").Zero()
	res, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return nil, err
	}

	result := make([]BigMapDiff, 0)
	arr := res.Get("aggregations.keys.buckets").Array()
	if int64(len(arr)) < offset {
		return nil, nil
	}

	if int64(len(arr)) < to {
		to = int64(len(arr))
	}

	arr = arr[offset:to]
	for _, item := range arr {
		bmd := item.Get("top_key.hits.hits.0")

		var b BigMapDiff
		b.ParseElasticJSON(bmd)
		b.Count = item.Get("doc_count").Int()
		result = append(result, b)
	}
	return result, nil
}

// GetBigMapDiffByPtrAndKeyHash -
func (e *Elastic) GetBigMapDiffByPtrAndKeyHash(address string, ptr int64, keyHash string, size, offset int64) ([]BigMapDiff, int64, error) {
	mustQuery := must(
		matchPhrase("address", address),
		matchPhrase("key_hash", keyHash),
	)
	if ptr != 0 {
		mustQuery.Extend(term("ptr", ptr))
	}
	b := boolQ(mustQuery)

	if ptr == 0 {
		b.Get("bool").Extend(
			notMust(
				qItem{
					"exists": qItem{
						"field": "ptr",
					},
				},
			),
		)
	}

	if size == 0 {
		size = defaultSize
	}

	query := newQuery().Query(b).Sort("level", "desc").Size(size).From(offset)
	res, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return nil, 0, err
	}

	result := make([]BigMapDiff, 0)
	for _, item := range res.Get("hits.hits").Array() {
		var b BigMapDiff
		b.ParseElasticJSON(item)
		result = append(result, b)
	}

	total := res.Get("hits.total.value").Int()
	return result, total, nil
}

// GetBigMapDiffsJSONByOperationID -
func (e *Elastic) GetBigMapDiffsJSONByOperationID(operationID string) ([]gjson.Result, error) {
	query := newQuery().
		Query(
			boolQ(
				must(
					matchPhrase("operation_id", operationID),
				),
			),
		).All()

	res, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return nil, err
	}
	return res.Get("hits.hits").Array(), nil
}

// GetAllBigMapDiff -
func (e *Elastic) GetAllBigMapDiff() ([]models.BigMapDiff, error) {
	bmd := make([]models.BigMapDiff, 0)

	result, err := e.createScroll(DocBigMapDiff, 1000, base{})
	if err != nil {
		return nil, err
	}
	for {
		scrollID := result.Get("_scroll_id").String()
		hits := result.Get("hits.hits")
		if hits.Get("#").Int() < 1 {
			break
		}

		for _, item := range hits.Array() {
			var c models.BigMapDiff
			c.ParseElasticJSON(item)
			bmd = append(bmd, c)
		}

		result, err = e.queryScroll(scrollID)
		if err != nil {
			return nil, err
		}
	}

	return bmd, nil
}
