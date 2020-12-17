package bulk

import (
	"bytes"
	"context"
	"encoding/json"
	stdJSON "encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

func (storage *Storage) bulk(buf *bytes.Buffer) error {
	req := esapi.BulkRequest{
		Body:    bytes.NewReader(buf.Bytes()),
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), storage.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var response core.BulkResponse
	err = storage.es.GetResponse(res, &response)
	return err
}

// Insert -
func (storage *Storage) Insert(items []models.Model) error {
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
			if err := storage.bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// Update -
func (storage *Storage) Update(updates []models.Model) error {
	if len(updates) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range updates {
		if _, err := bulk.WriteString(fmt.Sprintf(`{"update":{"_id":"%s","_index":"%s"}}`, updates[i].GetID(), updates[i].GetIndex())); err != nil {
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
			if err := storage.bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// Delete -
func (storage *Storage) Delete(updates []models.Model) error {
	if len(updates) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range updates {
		meta := []byte(fmt.Sprintf(`{"delete":{"_index":"%s","_id":"%s"}}%s`, updates[i].GetIndex(), updates[i].GetID(), "\n"))
		bulk.Grow(len(meta))
		bulk.Write(meta)

		if (i%1000 == 0 && i > 0) || i == len(updates)-1 {
			if err := storage.bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// RemoveField -
func (storage *Storage) RemoveField(script string, where []models.Model) error {
	if len(where) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range where {
		meta := fmt.Sprintf(`{ "update": { "_id": "%s", "_index": "%s"}}%s{"script" : "%s"}%s`, where[i].GetID(), where[i].GetIndex(), "\n", script, "\n")
		bulk.Grow(len(meta))
		bulk.WriteString(meta)

		if (i%1000 == 0 && i > 0) || i == len(where)-1 {
			if err := storage.bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// UpdateField -
func (storage *Storage) UpdateField(where []contract.Contract, fields ...string) error {
	if len(where) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range where {
		updated, err := storage.es.BuildFieldsForModel(where[i], fields...)
		if err != nil {
			return err
		}
		meta := fmt.Sprintf(`{ "update": { "_id": "%s", "_index": "%s"}}%s%s%s`, where[i].GetID(), where[i].GetIndex(), "\n", string(updated), "\n")
		bulk.Grow(len(meta))
		bulk.WriteString(meta)

		if (i%1000 == 0 && i > 0) || i == len(where)-1 {
			if err := storage.bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}
