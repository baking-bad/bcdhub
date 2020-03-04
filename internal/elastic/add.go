package elastic

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// AddDocument -
func (e *Elastic) AddDocument(v interface{}, index string) (s string, err error) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	req := esapi.IndexRequest{
		Index:   index,
		Body:    bytes.NewReader(b),
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), e)
	if err != nil {
		return
	}
	defer res.Body.Close()

	r, err := e.getResponse(res)
	if err != nil {
		return
	}
	return r.Get("_id").String(), nil
}

// AddDocumentWithID -
func (e *Elastic) AddDocumentWithID(v interface{}, index, docID string) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	req := esapi.IndexRequest{
		Index:      index,
		Body:       bytes.NewReader(b),
		Refresh:    "true",
		DocumentID: docID,
	}

	res, err := req.Do(context.Background(), e)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	r, err := e.getResponse(res)
	if err != nil {
		return "", err
	}
	return r.Get("_id").String(), nil
}
