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
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"match_phrase": map[string]interface{}{
				"operation_id": operationID,
			},
		},
		"size": 1000,
	}

	res, err := e.query(DocBigMapDiff, query)
	if err != nil {
		return *res, err
	}
	return res.Get("hits.hits.#._source"), nil
}

// GetBigMapDiffsByKeyHash -
func (e *Elastic) GetBigMapDiffsByKeyHash(keys []string, level int64) (gjson.Result, error) {
	must := make([]map[string]interface{}, len(keys))
	for i := range keys {
		must[i] = map[string]interface{}{
			"match_phrase": map[string]interface{}{
				"key_hash": keys[i],
			},
		}
	}
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": must,
				"filter": map[string]interface{}{
					"range": map[string]interface{}{
						"level": map[string]interface{}{
							"lt": level,
						},
					},
				},
			},
		},
		"sort": map[string]interface{}{
			"level": map[string]interface{}{
				"order": "desc",
			},
		},
		"size": 1000,
	}

	res, err := e.query(DocBigMapDiff, query)
	if err != nil {
		return *res, err
	}
	return res.Get("hits.hits.#._source"), nil
}

// GetBigMapDiffsForAddress -
func (e *Elastic) GetBigMapDiffsForAddress(address string) (gjson.Result, error) {
	query := map[string]interface{}{
		"size": 0,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"match_phrase": map[string]interface{}{
						"address": address,
					},
				},
			},
		},
		"aggs": map[string]interface{}{
			"group_by_hash": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "key_hash.keyword",
					"size":  10000,
				},
				"aggs": map[string]interface{}{
					"group_docs": map[string]interface{}{
						"top_hits": map[string]interface{}{
							"size": 1,
							"sort": map[string]interface{}{
								"level": map[string]interface{}{
									"order": "desc",
								},
							},
						},
					},
				},
			},
		},
	}
	res, err := e.query(DocBigMapDiff, query)
	if err != nil {
		return *res, err
	}
	return res.Get("aggregations.group_by_hash.buckets.#.group_docs.hits.hits.0._source"), nil
}
