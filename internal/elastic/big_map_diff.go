package elastic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/google/uuid"
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
