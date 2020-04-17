package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Identifiable -
type Identifiable interface {
	GetID() string
}

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

// BulkUpdate -
func (e *Elastic) BulkUpdate(index string, updates []Identifiable) error {
	bulk := bytes.NewBuffer([]byte{})
	for i := range updates {
		meta := []byte(fmt.Sprintf(`{ "update": { "_id": "%s"}}%s{ "doc": `, updates[i].GetID(), "\n"))
		data, err := json.Marshal(updates[i])
		if err != nil {
			log.Println(err)
			continue
		}
		data = append(data, "}\n"...)

		bulk.Grow(len(meta) + len(data))
		bulk.Write(meta)
		bulk.Write(data)
	}
	return e.Bulk(index, bulk)
}

// BulkDelete -
func (e *Elastic) BulkDelete(index string, updates []Identifiable) error {
	bulk := bytes.NewBuffer([]byte{})
	for i := range updates {
		meta := []byte(fmt.Sprintf(`{ "delete": { "_id": "%s"}}%s{ "doc": `, updates[i].GetID(), "\n"))
		data, err := json.Marshal(updates[i])
		if err != nil {
			log.Println(err)
			continue
		}
		data = append(data, "}\n"...)

		bulk.Grow(len(meta) + len(data))
		bulk.Write(meta)
		bulk.Write(data)
	}
	return e.Bulk(index, bulk)
}

// BulkInsertOperations -
func (e *Elastic) BulkInsertOperations(v []models.Operation) error {
	bulk := bytes.NewBuffer([]byte{})
	for i := range v {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id": "%s"} }%s`, v[i].ID, "\n"))
		data, err := json.Marshal(v[i])
		if err != nil {
			return err
		}
		data = append(data, "\n"...)

		bulk.Grow(len(meta) + len(data))
		bulk.Write(meta)
		bulk.Write(data)
	}
	return e.Bulk(DocOperations, bulk)
}

// BulkInsertContracts -
func (e *Elastic) BulkInsertContracts(v []models.Contract) error {
	bulk := bytes.NewBuffer([]byte{})
	for i := range v {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id": "%s"} }%s`, v[i].ID, "\n"))
		data, err := json.Marshal(v[i])
		if err != nil {
			return err
		}
		data = append(data, "\n"...)

		bulk.Grow(len(meta) + len(data))
		bulk.Write(meta)
		bulk.Write(data)
	}
	return e.Bulk(DocContracts, bulk)
}

// BulkSaveBigMapDiffs -
func (e *Elastic) BulkSaveBigMapDiffs(diffs []models.BigMapDiff) error {
	bulk := bytes.NewBuffer([]byte{})
	for i := range diffs {
		id := helpers.GenerateID()
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
	return e.Bulk(DocBigMapDiff, bulk)
}

// BulkInsertMigrations -
func (e *Elastic) BulkInsertMigrations(v []models.Migration) error {
	bulk := bytes.NewBuffer([]byte{})
	for i := range v {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id": "%s"} }%s`, v[i].ID, "\n"))
		data, err := json.Marshal(v[i])
		if err != nil {
			return err
		}
		data = append(data, "\n"...)

		bulk.Grow(len(meta) + len(data))
		bulk.Write(meta)
		bulk.Write(data)
	}
	return e.Bulk(DocMigrations, bulk)
}

// BulkInsertProtocols -
func (e *Elastic) BulkInsertProtocols(v []models.Protocol) error {
	bulk := bytes.NewBuffer([]byte{})
	for i := range v {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_id": "%s"} }%s`, v[i].ID, "\n"))
		data, err := json.Marshal(v[i])
		if err != nil {
			return err
		}
		data = append(data, "\n"...)

		bulk.Grow(len(meta) + len(data))
		bulk.Write(meta)
		bulk.Write(data)
	}
	return e.Bulk(DocProtocol, bulk)
}
