package elastic

import (
	"context"
	"fmt"
	"reflect"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/tidwall/gjson"
)

// GetByID -
func (e *Elastic) GetByID(ret Model) error {
	req := esapi.GetRequest{
		Index:      ret.GetIndex(),
		DocumentID: ret.GetID(),
	}
	resp, err := req.Do(context.Background(), e)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	result, err := e.getResponse(resp)
	if err != nil {
		return err
	}
	if !result.Get("found").Bool() {
		return fmt.Errorf("%s: %s %s", RecordNotFound, ret.GetIndex(), ret.GetID())
	}
	ret.ParseElasticJSON(result)
	return nil
}

// GetAll -
func (e *Elastic) GetAll(output interface{}) error {
	typ, err := getElementType(output)
	if err != nil {
		return err
	}
	index, err := getIndex(typ)
	if err != nil {
		return err
	}
	return e.getByScroll(index, nil, typ, output)
}

// GetByNetwork -
func (e *Elastic) GetByNetwork(network string, output interface{}) error {
	typ, err := getElementType(output)
	if err != nil {
		return err
	}
	index, err := getIndex(typ)
	if err != nil {
		return err
	}

	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("network", network),
			),
		),
	).Sort("level", "asc")
	return e.getByScroll(index, query, typ, output)
}

// GetByIDs -
func (e *Elastic) GetByIDs(ids []string, output interface{}) (err error) {
	typ, err := getElementType(output)
	if err != nil {
		return err
	}
	index, err := getIndex(typ)
	if err != nil {
		return err
	}

	query := newQuery().Query(
		qItem{
			"ids": qItem{
				"values": ids,
			},
		},
	)
	return e.getByScroll(index, query, typ, output)
}

func getElementType(output interface{}) (reflect.Type, error) {
	arr := reflect.TypeOf(output)
	if arr.Kind() != reflect.Ptr {
		return arr.Elem(), fmt.Errorf("Invalid `output` type: %s", arr.Kind())
	}
	return arr.Elem().Elem(), nil
}

func getIndex(typ reflect.Type) (string, error) {
	newItem := reflect.New(typ)
	interfaceType := reflect.TypeOf((*Model)(nil)).Elem()
	if !newItem.Type().Implements(interfaceType) {
		return "", fmt.Errorf("Implements: 'output' is not implemented `Model` interface")
	}

	getIndex := newItem.MethodByName("GetIndex")
	if !getIndex.IsValid() {
		return "", fmt.Errorf("getIndex: 'output' is not implemented `Model` interface")
	}
	getIndexResult := getIndex.Call(nil)
	if len(getIndexResult) != 1 {
		return "", fmt.Errorf("Something went wrong during call GetIndex")
	}
	return getIndexResult[0].Interface().(string), nil
}

func parseResponseItem(item gjson.Result, typ reflect.Type) (reflect.Value, error) {
	n := reflect.New(typ)
	parse := n.MethodByName("ParseElasticJSON")
	if !parse.IsValid() {
		return n.Elem(), fmt.Errorf("parse: 'output' is not implemented `Model` interface")
	}
	parse.Call([]reflect.Value{reflect.ValueOf(item)})
	return n.Elem(), nil
}

func (e *Elastic) getByScroll(index string, query base, typ reflect.Type, output interface{}) error {
	result, err := e.createScroll(index, 1000, query)
	if err != nil {
		return err
	}
	el := reflect.ValueOf(output).Elem()
	for {
		scrollID := result.Get("_scroll_id").String()
		hits := result.Get("hits.hits")
		if hits.Get("#").Int() < 1 {
			break
		}

		for _, item := range hits.Array() {
			n, err := parseResponseItem(item, typ)
			if err != nil {
				return err
			}
			el.Set(reflect.Append(el, n))
		}

		result, err = e.queryScroll(scrollID)
		if err != nil {
			return err
		}
	}
	return nil
}
