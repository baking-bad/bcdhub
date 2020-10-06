package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/tidwall/gjson"
)

// UpdateDoc - updates document
func (e *Elastic) UpdateDoc(model Model) (result gjson.Result, err error) {
	b, err := json.Marshal(model)
	if err != nil {
		return
	}
	req := esapi.IndexRequest{
		Index:      model.GetIndex(),
		DocumentID: model.GetID(),
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

// UpdateFields -
func (e *Elastic) UpdateFields(index, id string, data interface{}, fields ...string) error {
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

	b, err := json.Marshal(map[string]interface{}{
		"doc": updateFields,
	})
	if err != nil {
		return err
	}
	req := esapi.UpdateRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(b),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), e)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	_, err = e.getResponse(res)
	return err
}
