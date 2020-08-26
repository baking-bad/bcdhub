package elastic

import (
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

const defaultScrollSize = 1000

type scrollContext struct {
	Query base
	Size  int64

	e         *Elastic
	scrollIds map[string]struct{}
}

func newScrollContext(e *Elastic, query base, size int64) *scrollContext {
	return &scrollContext{
		e:         e,
		scrollIds: make(map[string]struct{}),

		Query: query,
		Size:  size,
	}
}

func (ctx *scrollContext) createScroll(index string, size int64, query map[string]interface{}) (result gjson.Result, err error) {
	var buf bytes.Buffer
	if err = json.NewEncoder(&buf).Encode(query); err != nil {
		return
	}

	var resp *esapi.Response
	options := []func(*esapi.SearchRequest){
		ctx.e.Search.WithContext(context.Background()),
		ctx.e.Search.WithIndex(index),
		ctx.e.Search.WithBody(&buf),
		ctx.e.Search.WithScroll(time.Minute),
		ctx.e.Search.WithSize(int(size)),
	}

	if resp, err = ctx.e.Search(
		options...,
	); err != nil {
		return
	}
	defer resp.Body.Close()

	return ctx.e.getResponse(resp)
}

func (ctx *scrollContext) queryScroll(scrollID string) (result gjson.Result, err error) {
	resp, err := ctx.e.Scroll(ctx.e.Scroll.WithScrollID(scrollID), ctx.e.Scroll.WithScroll(time.Minute))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	return ctx.e.getResponse(resp)
}

func (ctx *scrollContext) removeScroll(scrollIDs []string) error {
	if len(scrollIDs) == 0 {
		return nil
	}

	resp, err := ctx.e.ClearScroll(ctx.e.ClearScroll.WithScrollID(scrollIDs...))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (ctx *scrollContext) get(output interface{}) error {
	typ, err := getElementType(output)
	if err != nil {
		return err
	}
	index, err := getIndex(typ)
	if err != nil {
		return err
	}

	result, err := ctx.createScroll(index, ctx.Size, ctx.Query)
	if err != nil {
		return err
	}
	el := reflect.ValueOf(output).Elem()
	for {
		scrollID := result.Get("_scroll_id").String()
		ctx.scrollIds[scrollID] = struct{}{}

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

		result, err = ctx.queryScroll(scrollID)
		if err != nil {
			return err
		}
	}

	return ctx.clear()
}

func (ctx *scrollContext) clear() error {
	ctx.Query = nil
	ctx.Size = 0

	ids := make([]string, 0)
	for k := range ctx.scrollIds {
		ids = append(ids, k)
	}
	return ctx.removeScroll(ids)
}

func getElementType(output interface{}) (reflect.Type, error) {
	arr := reflect.TypeOf(output)
	if arr.Kind() != reflect.Ptr {
		return arr.Elem(), errors.Errorf("Invalid `output` type: %s", arr.Kind())
	}
	return arr.Elem().Elem(), nil
}

func getIndex(typ reflect.Type) (string, error) {
	newItem := reflect.New(typ)
	interfaceType := reflect.TypeOf((*Model)(nil)).Elem()
	if !newItem.Type().Implements(interfaceType) {
		return "", errors.Errorf("Implements: 'output' is not implemented `Model` interface")
	}

	getIndex := newItem.MethodByName("GetIndex")
	if !getIndex.IsValid() {
		return "", errors.Errorf("getIndex: 'output' is not implemented `Model` interface")
	}
	getIndexResult := getIndex.Call(nil)
	if len(getIndexResult) != 1 {
		return "", errors.Errorf("Something went wrong during call GetIndex")
	}
	return getIndexResult[0].Interface().(string), nil
}

func parseResponseItem(item gjson.Result, typ reflect.Type) (reflect.Value, error) {
	n := reflect.New(typ)
	parse := n.MethodByName("ParseElasticJSON")
	if !parse.IsValid() {
		return n.Elem(), errors.Errorf("parse: 'output' is not implemented `Model` interface")
	}
	parse.Call([]reflect.Value{reflect.ValueOf(item)})
	return n.Elem(), nil
}
