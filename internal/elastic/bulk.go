package elastic

import (
	"bytes"
	"context"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

func (e *Elastic) bulk(buf *bytes.Buffer) error {
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
	err = e.getResponse(res, &response)
	return err
}

// BulkInsert -
func (e *Elastic) BulkInsert(items []Model) error {
	if len(items) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range items {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id": "%s", "_index": "%s"} }%s`, items[i].GetID(), items[i].GetIndex(), "\n"))
		data, err := json.Marshal(items[i])
		if err != nil {
			return err
		}
		data = append(data, "\n"...)

		bulk.Grow(len(meta) + len(data))
		if _, err := bulk.Write(meta); err != nil {
			return err
		}
		if _, err := bulk.Write(data); err != nil {
			return err
		}

		if (i%1000 == 0 && i > 0) || i == len(items)-1 {
			if err := e.bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// BulkUpdate -
func (e *Elastic) BulkUpdate(updates []Model) error {
	if len(updates) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range updates {
		meta := []byte(fmt.Sprintf(`{ "update": { "_id": "%s", "_index": "%s"}}%s{ "doc": `, updates[i].GetID(), updates[i].GetIndex(), "\n"))
		data, err := json.Marshal(updates[i])
		if err != nil {
			return err
		}
		data = append(data, "}\n"...)

		bulk.Grow(len(meta) + len(data))
		bulk.Write(meta)
		bulk.Write(data)

		if (i%1000 == 0 && i > 0) || i == len(updates)-1 {
			if err := e.bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// BulkDelete -
func (e *Elastic) BulkDelete(updates []Model) error {
	if len(updates) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range updates {
		meta := []byte(fmt.Sprintf(`{ "delete": { "_index": "%s", "_id": "%s"}}%s`, updates[i].GetIndex(), updates[i].GetID(), "\n"))
		bulk.Grow(len(meta))
		bulk.Write(meta)

		if (i%1000 == 0 && i > 0) || i == len(updates)-1 {
			if err := e.bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// BulkRemoveField -
func (e *Elastic) BulkRemoveField(script string, where []Model) error {
	if len(where) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range where {
		meta := fmt.Sprintf(`{ "update": { "_id": "%s", "_index": "%s"}}%s{"script" : "%s"}%s`, where[i].GetID(), where[i].GetIndex(), "\n", script, "\n")
		bulk.Grow(len(meta))
		bulk.WriteString(meta)

		if (i%1000 == 0 && i > 0) || i == len(where)-1 {
			if err := e.bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

// BulkUpdateField -
func (e *Elastic) BulkUpdateField(where []models.Contract, fields ...string) error {
	if len(where) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range where {
		updated, err := e.buildFieldsForModel(where[i], fields...)
		if err != nil {
			return err
		}
		meta := fmt.Sprintf(`{ "update": { "_id": "%s", "_index": "%s"}}%s%s%s`, where[i].GetID(), where[i].GetIndex(), "\n", string(updated), "\n")
		bulk.Grow(len(meta))
		bulk.WriteString(meta)

		if (i%1000 == 0 && i > 0) || i == len(where)-1 {
			if err := e.bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}
