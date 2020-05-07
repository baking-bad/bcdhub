package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Bulk -
func (e *Elastic) Bulk(index string, buf *bytes.Buffer) error {
	req := esapi.BulkRequest{
		Index:   index,
		Body:    bytes.NewReader(buf.Bytes()),
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), e)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = e.getResponse(res)
	return err
}

// BulkInsert -
func (e *Elastic) BulkInsert(items []Model) error {
	if len(items) == 0 {
		return nil
	}
	index := items[0].GetIndex()
	bulk := bytes.NewBuffer([]byte{})
	for i := range items {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id": "%s"} }%s`, items[i].GetID(), "\n"))
		data, err := json.Marshal(items[i])
		if err != nil {
			return err
		}
		data = append(data, "\n"...)

		bulk.Grow(len(meta) + len(data))
		bulk.Write(meta)
		bulk.Write(data)
	}
	return e.Bulk(index, bulk)
}

// BulkUpdate -
func (e *Elastic) BulkUpdate(updates []Model) error {
	if len(updates) == 0 {
		return nil
	}
	index := updates[0].GetIndex()
	bulk := bytes.NewBuffer([]byte{})
	for i := range updates {
		meta := []byte(fmt.Sprintf(`{ "update": { "_id": "%s"}}%s{ "doc": `, updates[i].GetID(), "\n"))
		data, err := json.Marshal(updates[i])
		if err != nil {
			return err
		}
		data = append(data, "}\n"...)

		bulk.Grow(len(meta) + len(data))
		bulk.Write(meta)
		bulk.Write(data)
	}
	return e.Bulk(index, bulk)
}

// BulkDelete -
func (e *Elastic) BulkDelete(updates []Model) error {
	if len(updates) == 0 {
		return nil
	}
	index := updates[0].GetIndex()
	bulk := bytes.NewBuffer([]byte{})
	for i := range updates {
		meta := []byte(fmt.Sprintf(`{ "delete": { "_index": "%s", "_id": "%s"}}%s`, index, updates[i].GetID(), "\n"))
		bulk.Grow(len(meta))
		bulk.Write(meta)
	}
	return e.Bulk(index, bulk)
}

// BulkRemoveField -
func (e *Elastic) BulkRemoveField(script string, where []Model) error {
	if len(where) == 0 {
		return nil
	}
	index := where[0].GetIndex()
	bulk := bytes.NewBuffer([]byte{})
	for i := range where {
		meta := fmt.Sprintf(`{ "update": { "_id": "%s"}}%s{"script" : "%s"}%s`, where[i].GetID(), "\n", script, "\n")
		bulk.Grow(len(meta))
		bulk.WriteString(meta)
	}
	return e.Bulk(index, bulk)
}
