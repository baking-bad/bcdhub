package elastic

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// GetUniqueBigMapDiffsByOperationID -
func (e *Elastic) GetUniqueBigMapDiffsByOperationID(operationID string) ([]models.BigMapDiff, error) {
	query := newQuery().
		Query(
			boolQ(
				filter(
					matchPhrase("operation_id", operationID),
				),
			),
		).Add(aggs(
		"keys", qItem{
			"terms": qItem{
				"field": "key_hash.keyword",
				"size":  maxQuerySize,
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

// GetPrevBigMapDiffs -
func (e *Elastic) GetPrevBigMapDiffs(filters []models.BigMapDiff, indexedTime int64, address string) ([]models.BigMapDiff, error) {
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
		Add(aggs(
			"keys", qItem{
				"terms": qItem{
					"field": "key_hash.keyword",
					"size":  maxQuerySize,
				},
				"aggs": qItem{
					"top_key": topHits(1, "indexed_time", "desc"),
				},
			},
		)).
		Sort("indexed_time", "desc").Zero()

	res, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return nil, err
	}

	response := make([]models.BigMapDiff, 0)
	for _, item := range res.Get("aggregations.keys.buckets").Array() {
		bmd := item.Get("top_key.hits.hits.0")
		if bmd.Get("value").String() == "" {
			continue
		}
		var b models.BigMapDiff
		b.ParseElasticJSON(bmd)
		response = append(response, b)
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

// GetBigMapKeys -
func (e *Elastic) GetBigMapKeys(ptr int64, network, searchText string, size, offset int64) ([]BigMapDiff, error) {
	if ptr < 0 {
		return nil, fmt.Errorf("Invalid pointer value: %d", ptr)
	}

	mustQuery := []qItem{
		matchPhrase("network", network),
		term("ptr", ptr),
	}

	if searchText != "" {
		mustQuery = append(mustQuery, queryString(searchText, []string{"key", "key_hash", "key_strings", "bin_path"}))
	}

	b := boolQ(
		must(mustQuery...),
	)

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
func (e *Elastic) GetBigMapDiffByPtrAndKeyHash(ptr int64, network, keyHash string, size, offset int64) ([]BigMapDiff, int64, error) {
	if ptr < 0 {
		return nil, 0, fmt.Errorf("Invalid pointer value: %d", ptr)
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

// GetAllBigMapDiffByPtr -
func (e *Elastic) GetAllBigMapDiffByPtr(address, network string, ptr int64) ([]models.BigMapDiff, error) {
	bmd := make([]models.BigMapDiff, 0)

	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				matchPhrase("address", address),
				term("ptr", ptr),
			),
		),
	).Add(
		aggs("keys", qItem{
			"terms": qItem{
				"field": "key_hash.keyword",
			},
			"aggs": qItem{
				"top_key": topHits(1, "indexed_time", "desc"),
			},
		}),
	).Sort("indexed_time", "desc").Zero()

	result, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return nil, err
	}
	buckets := result.Get("aggregations.keys.buckets").Array()
	for _, item := range buckets {
		hit := item.Get("top_key.hits.hits.0")

		var b models.BigMapDiff
		b.ParseElasticJSON(hit)
		bmd = append(bmd, b)
	}
	return bmd, nil
}

// GetBigMapDiffsWithEmptyPtr -
func (e *Elastic) GetBigMapDiffsWithEmptyPtr() (response []models.BigMapDiff, err error) {
	query := newQuery().Query(
		boolQ(
			notMust(
				exists("ptr"),
			),
		),
	).Sort("indexed_time", "desc")

	err = e.GetAllByQuery(query, &response)
	return
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

	err = e.GetAllByQuery(query, &response)
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

	err = e.GetAllByQuery(query, &response)
	return
}
