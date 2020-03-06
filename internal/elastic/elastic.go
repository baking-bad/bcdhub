package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/aopoltorzhicky/bcdhub/internal/logger"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/tidwall/gjson"
)

// Elastic -
type Elastic struct {
	*elasticsearch.Client
}

// New -
func New(addresses []string) (*Elastic, error) {
	elasticConfig := elasticsearch.Config{
		Addresses: addresses,
	}
	es, err := elasticsearch.NewClient(elasticConfig)
	if err != nil {
		return nil, err
	}
	e := &Elastic{es}
	r, err := e.TestConnection()
	if err != nil {
		return nil, err
	}
	logger.Info("Elasticsearch Server: %s", r.Get("version.number").String())

	return e, nil
}

// WaitNew -
func WaitNew(addresses []string) *Elastic {
	var es *Elastic
	var err error

	for es == nil {
		es, err = New(addresses)
		if err != nil {
			logger.Warning("Waiting elastic up 30 second...")
			time.Sleep(time.Second * 30)
		}
	}
	return es
}

func (e *Elastic) getResponse(resp *esapi.Response) (result gjson.Result, err error) {
	if resp.IsError() {
		return result, fmt.Errorf(resp.String())
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	result = gjson.ParseBytes(b)
	return
}

func (e *Elastic) query(indices []string, query map[string]interface{}, source ...string) (result gjson.Result, err error) {
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return
	}

	// log.Print(buf.String())

	// Perform the search request.
	var resp *esapi.Response

	options := []func(*esapi.SearchRequest){
		e.Search.WithContext(context.Background()),
		e.Search.WithIndex(indices...),
		e.Search.WithBody(&buf),
		e.Search.WithSource(source...),
	}

	if resp, err = e.Search(
		options...,
	); err != nil {
		return
	}

	defer resp.Body.Close()

	return e.getResponse(resp)
}

func (e *Elastic) executeSQL(sqlString string) (result gjson.Result, err error) {
	query := qItem{
		"query": sqlString,
	}

	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return
	}

	options := []func(*esapi.SQLQueryRequest){
		e.SQL.Query.WithContext(context.Background()),
	}

	var resp *esapi.Response
	if resp, err = e.SQL.Query(&buf, options...); err != nil {
		return
	}
	defer resp.Body.Close()

	return e.getResponse(resp)
}

func (e *Elastic) createScroll(index string, size int64, query map[string]interface{}) (result gjson.Result, err error) {
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return
	}

	// log.Print(buf.String())

	var resp *esapi.Response
	options := []func(*esapi.SearchRequest){
		e.Search.WithContext(context.Background()),
		e.Search.WithIndex(index),
		e.Search.WithBody(&buf),
		e.Search.WithScroll(time.Minute),
		e.Search.WithSize(int(size)),
	}

	if resp, err = e.Search(
		options...,
	); err != nil {
		return
	}
	defer resp.Body.Close()

	return e.getResponse(resp)
}

func (e *Elastic) queryScroll(scrollID string) (result gjson.Result, err error) {
	resp, err := e.Scroll(e.Scroll.WithScrollID(scrollID), e.Scroll.WithScroll(time.Minute))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return e.getResponse(resp)
}

// TestConnection -
func (e *Elastic) TestConnection() (result gjson.Result, err error) {
	res, err := e.Info()
	if err != nil {
		return
	}

	return e.getResponse(res)
}

// UpdateDoc - updates document by ID
func (e *Elastic) UpdateDoc(index, id string, v interface{}) (result gjson.Result, err error) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(b),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), e)
	if err != nil {
		return
	}
	defer res.Body.Close()

	result, err = e.getResponse(res)
	return
}

// CreateIndexIfNotExists -
func (e *Elastic) CreateIndexIfNotExists(index string) error {
	req := esapi.IndicesExistsRequest{
		Index: []string{index},
	}
	res, err := req.Do(context.Background(), e)
	if err != nil {
		return err
	}

	if !res.IsError() {
		return nil
	}

	jsonFile, err := os.Open(fmt.Sprintf("mappings/%s.json", index))
	if err != nil {
		res, err = e.Indices.Create(index)
		if err != nil {
			return err
		}
		if res.IsError() {
			return fmt.Errorf("%s", res)
		}
		return nil
	}
	defer jsonFile.Close()

	res, err = e.Indices.Create(index, e.Indices.Create.WithBody(jsonFile))
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("%s", res)
	}
	return nil
}
