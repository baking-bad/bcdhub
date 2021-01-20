package core

import (
	"bytes"
	"context"
	stdJSON "encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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
	return e, e.TestConnection()
}

// WaitNew -
func WaitNew(addresses []string, timeout int) *Elastic {
	var es *Elastic
	var err error

	for es == nil {
		es, err = New(addresses)
		if err != nil {
			logger.Warning("Waiting elastic up %d seconds...", timeout)
			time.Sleep(time.Second * time.Duration(timeout))
		}
	}
	return es
}

// GetAPI -
func (e *Elastic) GetAPI() *esapi.API {
	return e.API
}

// GetResponse -
func (e *Elastic) GetResponse(resp *esapi.Response, result interface{}) error {
	if resp.IsError() {
		if resp.StatusCode == 404 {
			return NewRecordNotFoundErrorFromResponse(resp)
		}
		return errors.Errorf(resp.String())
	}

	if result == nil {
		return nil
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, result)
}

func (e *Elastic) getTextResponse(resp *esapi.Response) (string, error) {
	if resp.IsError() {
		return "", errors.Errorf(resp.String())
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// Query -
func (e *Elastic) Query(indices []string, query map[string]interface{}, response interface{}, source ...string) (err error) {
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return
	}

	// logger.InterfaceToJSON(query)
	// logger.InterfaceToJSON(indices)

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

	return e.GetResponse(resp, response)
}

// ExecuteSQL -
func (e *Elastic) ExecuteSQL(sqlString string, response interface{}) (err error) {
	query := Item{
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

	return e.GetResponse(resp, response)
}

// TestConnection -
func (e *Elastic) TestConnection() (err error) {
	res, err := e.Info()
	if err != nil {
		return
	}

	var result TestConnectionResponse
	return e.GetResponse(res, &result)
}

func (e *Elastic) createIndexIfNotExists(index string) error {
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
			return errors.Errorf("%s", res)
		}
		return nil
	}
	defer jsonFile.Close()

	res, err = e.Indices.Create(index, e.Indices.Create.WithBody(jsonFile))
	if err != nil {
		return err
	}
	if res.IsError() {
		return errors.Errorf("%s", res)
	}
	return nil
}

// CreateIndexes -
func (e *Elastic) CreateIndexes() error {
	for _, index := range models.AllDocuments() {
		if err := e.createIndexIfNotExists(index); err != nil {
			return err
		}
	}
	return nil
}

// UpdateByQueryScript -
//nolint
func (e *Elastic) UpdateByQueryScript(indices []string, query map[string]interface{}, source ...string) (err error) {
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return
	}

	// logger.InterfaceToJSON(query)
	// logger.InterfaceToJSON(indices)

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

	var v interface{}
	return e.GetResponse(resp, &v)
}

func (e *Elastic) deleteByQuery(indices []string, query map[string]interface{}) (result *DeleteByQueryResponse, err error) {
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return
	}

	// logger.InterfaceToJSON(query)
	// logger.InterfaceToJSON(indices)

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
		return
	}

	defer resp.Body.Close()

	err = e.GetResponse(resp, &result)
	return
}

// DeleteByLevelAndNetwork -
func (e *Elastic) DeleteByLevelAndNetwork(indices []string, network string, maxLevel int64) error {
	query := NewQuery().Query(
		Bool(
			Filter(
				Match("network", network),
				Range("level", Item{"gt": maxLevel}),
			),
		),
	)
	end := false

	for !end {
		response, err := e.deleteByQuery(indices, query)
		if err != nil {
			return err
		}

		end = response.VersionConflicts == 0
		logger.Info("Removed %d/%d records from %s", response.Deleted, response.Total, strings.Join(indices, ","))
	}
	return nil
}

// DeleteIndices -
func (e *Elastic) DeleteIndices(indices []string) error {
	options := []func(*esapi.IndicesDeleteRequest){
		e.Indices.Delete.WithAllowNoIndices(true),
	}

	resp, err := e.Indices.Delete(indices, options...)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.IsError() {
		return errors.Errorf(resp.Status())
	}

	return nil
}

// ReloadSecureSettings -
func (e *Elastic) ReloadSecureSettings() error {
	resp, err := e.Nodes.ReloadSecureSettings()
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.IsError() {
		return errors.Errorf(resp.Status())
	}

	return nil
}

// DeleteByContract -
// TODO - delete context
func (e *Elastic) DeleteByContract(indices []string, network, address string) error {
	filters := make([]Item, 0)
	if network != "" {
		filters = append(filters, Match("network", network))
	}
	if address != "" {
		filters = append(filters, MatchPhrase("contract", address))
	}
	query := NewQuery().Query(
		Bool(
			Filter(filters...),
		),
	)
	end := false

	for !end {
		response, err := e.deleteByQuery(indices, query)
		if err != nil {
			return err
		}

		end = response.VersionConflicts == 0
		logger.Info("Removed %d/%d records from %s", response.Deleted, response.Total, strings.Join(indices, ","))
	}

	return nil
}

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
	return e.GetResponse(res, &response)
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
