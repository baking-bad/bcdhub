package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
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

func (e *Elastic) getTextResponse(resp *esapi.Response) (string, error) {
	if resp.IsError() {
		return "", fmt.Errorf(resp.String())
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (e *Elastic) query(indices []string, query map[string]interface{}, source ...string) (result gjson.Result, err error) {
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return
	}

	// var pretty bytes.Buffer
	// json.Indent(&pretty, buf.Bytes(), "", "  ")
	// log.Println(indices)
	// log.Println(pretty.String())

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

// CreateIndexes -
func (e *Elastic) CreateIndexes() error {
	for _, index := range []string{
		DocContracts,
		DocMetadata,
		DocBigMapDiff,
		DocOperations,
		DocMigrations,
		DocProtocol,
		DocBlocks,
	} {
		if err := e.CreateIndexIfNotExists(index); err != nil {
			return err
		}
	}
	return nil
}

func (e *Elastic) updateByQuery(indices []string, query map[string]interface{}, source ...string) (result gjson.Result, err error) {
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return
	}

	// var pretty bytes.Buffer
	// json.Indent(&pretty, buf.Bytes(), "", "  ")
	// log.Print(pretty.String())

	// Perform the update by query request.
	var resp *esapi.Response

	options := []func(*esapi.UpdateByQueryRequest){
		e.UpdateByQuery.WithContext(context.Background()),
		e.UpdateByQuery.WithBody(&buf),
		e.UpdateByQuery.WithSource(source...),
		e.UpdateByQuery.WithConflicts("proceed"),
	}

	if resp, err = e.UpdateByQuery(
		indices,
		options...,
	); err != nil {
		return
	}

	defer resp.Body.Close()

	return e.getResponse(resp)
}

func (e *Elastic) deleteByQuery(indices []string, query map[string]interface{}) (gjson.Result, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return gjson.Result{}, err
	}
	// var pretty bytes.Buffer
	// json.Indent(&pretty, buf.Bytes(), "", "  ")
	// log.Print(pretty.String())

	options := []func(*esapi.DeleteByQueryRequest){
		e.DeleteByQuery.WithContext(context.Background()),
		e.DeleteByQuery.WithConflicts("proceed"),
	}
	resp, err := e.DeleteByQuery(
		indices,
		bytes.NewReader(buf.Bytes()),
		options...,
	)
	if err != nil {
		return gjson.Result{}, err
	}

	defer resp.Body.Close()

	return e.getResponse(resp)
}

// DeleteByLevelAndNetwork -
func (e *Elastic) DeleteByLevelAndNetwork(indices []string, network string, maxLevel int64) error {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				rangeQ("level", qItem{"gt": maxLevel}),
			),
		),
	)
	end := false

	for !end {
		response, err := e.deleteByQuery(indices, query)
		if err != nil {
			return err
		}

		end = response.Get("version_conflicts").Int() == 0
		log.Printf("Removed %d/%d records from %s", response.Get("deleted").Int(), response.Get("total").Int(), strings.Join(indices, ","))
	}
	return nil
}
