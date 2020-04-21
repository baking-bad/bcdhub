package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// CreateAWSRepository -
func (e *Elastic) CreateAWSRepository(name, awsBucketName, awsRegion string) error {
	query := map[string]interface{}{
		"type": "s3",
		"settings": map[string]interface{}{
			"bucket":      awsBucketName,
			"endpoint":    fmt.Sprintf("s3.%s.amazonaws.com", awsRegion),
			"compress":    "true",
			"max_retries": 3,
		},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return err
	}

	options := []func(*esapi.SnapshotCreateRepositoryRequest){
		e.Snapshot.CreateRepository.WithContext(context.Background()),
		e.Snapshot.CreateRepository.WithVerify(false),
	}
	resp, err := e.Snapshot.CreateRepository(
		name,
		bytes.NewReader(buf.Bytes()),
		options...,
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = e.getResponse(resp)
	return err
}

// CreateSnapshots -
func (e *Elastic) CreateSnapshots(repository, snapshot string, indices []string) error {
	query := map[string]interface{}{
		"indices":              strings.Join(indices, ","),
		"ignore_unavailable":   true,
		"include_global_state": false,
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(query); err != nil {
		return err
	}
	options := []func(*esapi.SnapshotCreateRequest){
		e.Snapshot.Create.WithContext(context.Background()),
		e.Snapshot.Create.WithWaitForCompletion(true),
		e.Snapshot.Create.WithBody(bytes.NewReader(body.Bytes())),
	}
	resp, err := e.Snapshot.Create(
		repository,
		snapshot,
		options...,
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = e.getResponse(resp)
	return err
}

// RestoreSnapshots -
func (e *Elastic) RestoreSnapshots(repository, snapshot string, indices []string) error {
	query := map[string]interface{}{
		"indices":              strings.Join(indices, ","),
		"ignore_unavailable":   true,
		"include_global_state": false,
	}

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(query); err != nil {
		return err
	}
	options := []func(*esapi.SnapshotRestoreRequest){
		e.Snapshot.Restore.WithContext(context.Background()),
		e.Snapshot.Restore.WithWaitForCompletion(true),
		e.Snapshot.Restore.WithBody(bytes.NewReader(body.Bytes())),
	}
	resp, err := e.Snapshot.Restore(
		repository,
		snapshot,
		options...,
	)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = e.getResponse(resp)
	return err
}

// GetMappings -
func (e *Elastic) GetMappings(indices []string) (map[string]string, error) {
	options := []func(*esapi.IndicesGetMappingRequest){
		e.Indices.GetMapping.WithContext(context.Background()),
		e.Indices.GetMapping.WithIndex(indices...),
	}
	resp, err := e.Indices.GetMapping(
		options...,
	)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	response, err := e.getResponse(resp)
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for k, v := range response.Map() {
		result[k] = v.String()
	}
	return result, nil
}

// CreateMapping -
func (e *Elastic) CreateMapping(index string, r io.Reader) error {
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

	res, err = e.Indices.Create(index, e.Indices.Create.WithBody(r))
	if err != nil {
		return err
	}
	if res.IsError() {
		return fmt.Errorf("%s", res)
	}
	return nil
}
