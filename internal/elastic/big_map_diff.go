package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

// BulkSaveBigMapDiffs -
func (e *Elastic) BulkSaveBigMapDiffs(diffs []models.BigMapDiff) error {
	bulk := bytes.NewBuffer([]byte{})
	for i := range diffs {
		id := uuid.New().String()
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id": "%s"} }%s`, id, "\n"))
		data, err := json.Marshal(diffs[i])
		if err != nil {
			log.Println(err)
			continue
		}
		data = append(data, "\n"...)

		bulk.Grow(len(meta) + len(data))
		bulk.Write(meta)
		bulk.Write(data)
	}
	return e.BulkInsert(DocBigMapDiff, bulk)
}

// GetBigMapDiffsByOperationID -
func (e *Elastic) GetBigMapDiffsByOperationID(operationID string) (gjson.Result, error) {
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
		return res, err
	}
	return res.Get("hits.hits.#._source"), nil
}

// GetBigMapDiffsByKeyHash -
func (e *Elastic) GetBigMapDiffsByKeyHash(keys []string, level int64, address string) (gjson.Result, error) {
	shouldData := make([]qItem, len(keys))
	for i := range keys {
		shouldData[i] = matchPhrase("key_hash", keys[i])
	}

	b := boolQ(
		should(shouldData...),
		must(matchPhrase("address", address)),
		filter(
			rangeQ("level", qItem{"lt": level}),
		),
	)
	b.Get("bool").Append("minimum_should_match", 1)

	query := newQuery().Query(b).
		Add(qItem{
			"aggs": qItem{
				"last": qItem{
					"terms": qItem{
						"field": "key_hash.keyword",
					},
					"aggs": qItem{
						"bmd": topHits(1, "level", "desc"),
					},
				},
			},
		}).
		Zero()

	res, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return res, err
	}
	return res.Get("aggregations.last.buckets.#.bmd.hits.hits.0._source"), nil
}

// GetBigMapDiffsForAddress -
func (e *Elastic) GetBigMapDiffsForAddress(address string) (gjson.Result, error) {
	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("address", address),
			),
		),
	).Add(aggs(
		"group_by_hash", qItem{
			"terms": qItem{
				"field": "key_hash.keyword",
				"size":  maxQuerySize,
			},
			"aggs": qItem{
				"group_docs": topHits(1, "level", "desc"),
			},
		},
	)).Zero()

	res, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return res, err
	}
	return res.Get("aggregations.group_by_hash.buckets.#.group_docs.hits.hits.0._source"), nil
}

// GetBigMap -
func (e *Elastic) GetBigMap(address string, ptr int64) ([]BigMapDiff, error) {
	mustQuery := must(matchPhrase("address", address))
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

	query := newQuery().Query(b).Add(
		aggs("keys", qItem{
			"terms": qItem{
				"field": "key_hash.keyword",
				"size":  maxQuerySize,
				"order": qItem{
					"bucketsSort": "desc",
				},
			},
			"aggs": qItem{
				"top_key":     topHits(1, "level", "desc"),
				"bucketsSort": max("level"),
			},
		}),
	).Sort("level", "desc").Zero()
	res, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return nil, err
	}

	result := make([]BigMapDiff, 0)
	for _, item := range res.Get("aggregations.keys.buckets").Array() {
		bmd := item.Get("top_key.hits.hits.0")

		var b BigMapDiff
		b.ParseElasticJSON(bmd)
		b.Count = item.Get("doc_count").Int()
		result = append(result, b)
	}
	return result, nil
}

// GetBigMapDiffByPtrAndKeyHash -
func (e *Elastic) GetBigMapDiffByPtrAndKeyHash(address string, ptr int64, keyHash string) ([]models.BigMapDiff, error) {
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

	query := newQuery().Query(b).Sort("level", "desc").All()
	res, err := e.query([]string{DocBigMapDiff}, query)
	if err != nil {
		return nil, err
	}

	result := make([]models.BigMapDiff, 0)
	for _, item := range res.Get("hits.hits").Array() {
		var b models.BigMapDiff
		b.ParseElasticJSON(item)
		result = append(result, b)
	}
	return result, nil
}
