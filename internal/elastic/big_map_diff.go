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

	res, err := e.query(DocBigMapDiff, query)
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

	res, err := e.query(DocBigMapDiff, query)
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

	res, err := e.query(DocBigMapDiff, query)
	if err != nil {
		return res, err
	}
	return res.Get("aggregations.group_by_hash.buckets.#.group_docs.hits.hits.0._source"), nil
}
