package core

import (
	"bytes"
	"context"
	"reflect"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// UpdateDoc - updates document
func (e *Elastic) UpdateDoc(model models.Model) error {
	b, err := json.Marshal(model)
	if err != nil {
		return err
	}
	req := esapi.IndexRequest{
		Index:      model.GetIndex(),
		DocumentID: model.GetID(),
		Body:       bytes.NewReader(b),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), e)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return e.GetResponse(res, nil)
}

// BuildFieldsForModel -
func (e *Elastic) BuildFieldsForModel(data interface{}, fields ...string) ([]byte, error) {
	t := reflect.TypeOf(data)
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	mapFields := make(map[string]struct{})
	for i := range fields {
		mapFields[fields[i]] = struct{}{}
	}

	updateFields := make(map[string]interface{})
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := val.Field(i)
		if _, ok := mapFields[field.Name]; !ok {
			continue
		}
		tag := field.Tag.Get("json")
		tagName := strings.Split(tag, ",")[0]
		updateFields[tagName] = value.Interface()
	}

	return json.Marshal(map[string]interface{}{
		"doc":           updateFields,
		"doc_as_upsert": true,
	})
}

// UpdateFields -
func (e *Elastic) UpdateFields(index, id string, data interface{}, fields ...string) error {
	updated, err := e.BuildFieldsForModel(data, fields...)
	if err != nil {
		return err
	}

	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(updated),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), e)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return e.GetResponse(res, nil)
}
