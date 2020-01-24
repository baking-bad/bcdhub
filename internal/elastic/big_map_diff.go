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
func (e *Elastic) GetBigMapDiffsByOperationID(operationID string) (*gjson.Result, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": map[string]interface{}{
					"match": map[string]interface{}{
						"operation_id": operationID,
					},
				},
			},
		},
	}

	return e.query(DocBigMapDiff, query)
}

func parseBigMapDiff(data gjson.Result) models.BigMapDiff {
	return models.BigMapDiff{
		BinPath:     data.Get("_source.bin_path").String(),
		Ptr:         data.Get("_source.ptr").Int(),
		Key:         data.Get("_source.key").Value(),
		Value:       data.Get("_source.value").String(),
		KeyHash:     data.Get("_source.key_hash").String(),
		OperationID: data.Get("_source.operation_id").String(),
	}
}
