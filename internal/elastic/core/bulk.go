package core

import (
	"bytes"
	"context"
	stdJSON "encoding/json"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"
)

// Bulk -
func (e *Elastic) Bulk(buf *bytes.Buffer) error {
	req := esapi.BulkRequest{
		Body:    bytes.NewReader(buf.Bytes()),
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), e)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var response BulkResponse
	if err := e.GetResponse(res, &response); err != nil {
		return err
	}
	if response.Errors {
		return errors.Errorf("Bulk error: %s", string(response.Items))
	}
	return nil
}

// BulkInsert -
func (e *Elastic) BulkInsert(items []models.Model) error {
	if len(items) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range items {
		meta := fmt.Sprintf(`{"index":{"_id":"%s","_index":"%s"}}`, items[i].GetID(), items[i].GetIndex())
		if _, err := bulk.WriteString(meta); err != nil {
			return err
		}

		if err := bulk.WriteByte('\n'); err != nil {
			return err
		}

		data, err := json.Marshal(items[i])
		if err != nil {
			return err
		}

		if err := stdJSON.Compact(bulk, data); err != nil {
			return err
		}
		if err := bulk.WriteByte('\n'); err != nil {
			return err
		}

		if (i%1000 == 0 && i > 0) || i == len(items)-1 {
			if err := e.Bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// BulkUpdate -
func (e *Elastic) BulkUpdate(updates []models.Model) error {
	if len(updates) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range updates {
		if _, err := bulk.WriteString(fmt.Sprintf(`{"update":{"_id":"%s","_index":"%s", "retry_on_conflict": 2}}`, updates[i].GetID(), updates[i].GetIndex())); err != nil {
			return err
		}
		if err := bulk.WriteByte('\n'); err != nil {
			return err
		}
		data, err := json.Marshal(map[string]models.Model{
			"doc": updates[i],
		})
		if err != nil {
			return err
		}
		if err := stdJSON.Compact(bulk, data); err != nil {
			return err
		}
		if err := bulk.WriteByte('\n'); err != nil {
			return err
		}

		if (i%1000 == 0 && i > 0) || i == len(updates)-1 {
			if err := e.Bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// BulkDelete -
func (e *Elastic) BulkDelete(updates []models.Model) error {
	if len(updates) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range updates {
		meta := []byte(fmt.Sprintf(`{"delete":{"_index":"%s","_id":"%s"}}%s`, updates[i].GetIndex(), updates[i].GetID(), "\n"))
		bulk.Grow(len(meta))
		bulk.Write(meta)

		if (i%1000 == 0 && i > 0) || i == len(updates)-1 {
			if err := e.Bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// BulkRemoveField -
func (e *Elastic) BulkRemoveField(field string, where []models.Model) error {
	if len(where) == 0 {
		return nil
	}
	var sb strings.Builder
	if _, err := sb.WriteString("ctx._source."); err != nil {
		return err
	}

	idx := strings.LastIndex(field, ".")
	if idx > -1 {
		if _, err := sb.WriteString(field[:idx]); err != nil {
			return err
		}
		if _, err := sb.WriteString(fmt.Sprintf(`.remove('%s')`, field[idx+1:])); err != nil {
			return err
		}
	} else if _, err := sb.WriteString(fmt.Sprintf(`remove('%s')`, field)); err != nil {
		return err
	}

	bulk := bytes.NewBuffer([]byte{})
	for i := range where {
		meta := fmt.Sprintf(`{ "update": { "_id": "%s", "_index": "%s"}}%s{"script" : "%s"}%s`, where[i].GetID(), where[i].GetIndex(), "\n", sb.String(), "\n")
		bulk.Grow(len(meta))
		bulk.WriteString(meta)

		if (i%1000 == 0 && i > 0) || i == len(where)-1 {
			if err := e.Bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}
