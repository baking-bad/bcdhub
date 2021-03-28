package elastic

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/search"
	"github.com/cenkalti/backoff"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Elastic -
type Elastic struct {
	*elasticsearch.Client

	MaxPageSize int64
}

// New -
func New(addresses []string, maxPageSize int64) (*Elastic, error) {
	retryBackoff := backoff.NewExponentialBackOff()
	elasticConfig := elasticsearch.Config{
		Addresses:     addresses,
		RetryOnStatus: []int{502, 503, 504, 429},
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},
		MaxRetries: 5,
	}
	es, err := elasticsearch.NewClient(elasticConfig)
	if err != nil {
		return nil, err
	}

	if maxPageSize <= 0 {
		maxPageSize = consts.DefaultSize
	}

	e := &Elastic{es, maxPageSize}
	return e, e.ping()
}

// WaitNew -
func WaitNew(addresses []string, timeout int, maxPageSize int64) *Elastic {
	var es *Elastic
	var err error

	for es == nil {
		es, err = New(addresses, maxPageSize)
		if err != nil {
			logger.Warning("Waiting elastic up %d seconds...", timeout)
			time.Sleep(time.Second * time.Duration(timeout))
		}
	}
	return es
}

func (e *Elastic) getResponse(resp *esapi.Response, result interface{}) error {
	if resp.IsError() {
		if resp.StatusCode == 404 {
			return NewRecordNotFoundErrorFromResponse(resp)
		}
		return errors.Errorf(resp.String())
	}

	if result == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

func (e *Elastic) getTextResponse(resp *esapi.Response) (string, error) {
	if resp.IsError() {
		return "", errors.Errorf(resp.String())
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(resp.Body)
	return buf.String(), err
}

func (e *Elastic) query(indices []string, query map[string]interface{}, response interface{}, source ...string) (err error) {
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

	resp, err = e.Search(
		options...,
	)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return e.getResponse(resp, response)
}

func (e *Elastic) ping() (err error) {
	res, err := e.Info()
	if err != nil {
		return
	}
	defer res.Body.Close()

	var result TestConnectionResponse
	return e.getResponse(res, &result)
}

func (e *Elastic) createIndexIfNotExists(index string) error {
	req := esapi.IndicesExistsRequest{
		Index: []string{index},
	}
	res, err := req.Do(context.Background(), e)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if !res.IsError() {
		return nil
	}

	jsonFile, err := os.Open(fmt.Sprintf("mappings/%s.json", index))
	if err != nil {
		res, err = e.Indices.Create(index)
		if err != nil {
			return err
		}
		defer res.Body.Close()

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
	defer res.Body.Close()

	if res.IsError() {
		return errors.Errorf("%s", res)
	}
	return nil
}

// CreateIndexes -
func (e *Elastic) CreateIndexes() error {
	for _, index := range search.Indices {
		if err := e.createIndexIfNotExists(index); err != nil {
			return err
		}
	}
	return nil
}

func (e *Elastic) deleteWithQuery(indices []string, query map[string]interface{}) (result *DeleteByQueryResponse, err error) {
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return
	}

	// logger.InterfaceToJSON(query)
	// logger.InterfaceToJSON(indices)

	options := []func(*esapi.DeleteByQueryRequest){
		e.DeleteByQuery.WithContext(context.Background()),
		e.DeleteByQuery.WithConflicts("proceed"),
		e.DeleteByQuery.WithWaitForCompletion(true),
		e.DeleteByQuery.WithRefresh(true),
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

	err = e.getResponse(resp, &result)
	return
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
