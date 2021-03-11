package core

import (
	"bytes"
	"context"
	"reflect"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"
)

const defaultScrollSize = 1000

// ScrollContext -
type ScrollContext struct {
	Query     Base
	Size      int64
	ChunkSize int64

	Offset int64

	e         *Elastic
	scrollIds map[string]struct{}
}

// NewScrollContext -
func NewScrollContext(e *Elastic, query Base, size, chunkSize int64) *ScrollContext {
	if chunkSize == 0 {
		chunkSize = defaultScrollSize
	}
	return &ScrollContext{
		e:         e,
		scrollIds: make(map[string]struct{}),

		Query:     query,
		Size:      size,
		ChunkSize: chunkSize,
	}
}

func (ctx *ScrollContext) createScroll(index string, query map[string]interface{}) (response SearchResponse, err error) {
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
		ctx.e.Search.WithSize(int(ctx.ChunkSize)),
	}

	if resp, err = ctx.e.Search(
		options...,
	); err != nil {
		return
	}
	defer resp.Body.Close()

	err = ctx.e.GetResponse(resp, &response)
	return
}

func (ctx *ScrollContext) queryScroll(scrollID string) (response SearchResponse, err error) {
	resp, err := ctx.e.Scroll(ctx.e.Scroll.WithScrollID(scrollID), ctx.e.Scroll.WithScroll(time.Minute))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	err = ctx.e.GetResponse(resp, &response)
	return
}

func (ctx *ScrollContext) removeScroll(scrollIDs []string) error {
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

// Get -
func (ctx *ScrollContext) Get(output interface{}) error {
	typ, err := getElementType(output)
	if err != nil {
		return err
	}
	index, err := getIndex(typ)
	if err != nil {
		return err
	}

	result, err := ctx.createScroll(index, ctx.Query)
	if err != nil {
		return err
	}
	el := reflect.ValueOf(output).Elem()
	var count int64
	for {
		ctx.scrollIds[result.ScrollID] = struct{}{}

		hits := result.Hits.Hits
		if len(hits) < 1 {
			break
		}

		if ctx.Offset > 0 {
			if count+ctx.Size < ctx.Offset {
				count += int64(len(hits))
				result, err = ctx.queryScroll(result.ScrollID)
				if err != nil {
					return err
				}
				continue
			}

			length := int64(len(hits))
			offset := ctx.Offset - count
			if offset > 0 {
				hits = hits[offset:]
			}
			count += length

			if length < offset {
				break
			}
		}

		for _, item := range hits {
			n, err := parseResponseItem(item, typ)
			if err != nil {
				return err
			}

			if el.Kind() == reflect.Slice {
				el.Set(reflect.Append(el, n))
				if ctx.Size > 0 && el.Len() == int(ctx.Size) {
					return ctx.clear()
				}
			} else {
				el.Set(n)
			}
		}

		result, err = ctx.queryScroll(result.ScrollID)
		if err != nil {
			return err
		}
	}

	return ctx.clear()
}

func (ctx *ScrollContext) clear() error {
	ctx.Query = nil
	ctx.Size = 0
	ctx.ChunkSize = 0

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
	arr = arr.Elem()
	if arr.Kind() == reflect.Slice {
		return arr.Elem(), nil
	}
	return arr, nil
}

func getIndex(typ reflect.Type) (string, error) {
	newItem := reflect.New(typ)
	interfaceType := reflect.TypeOf((*models.Model)(nil)).Elem()

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

func parseResponseItem(hit Hit, typ reflect.Type) (reflect.Value, error) {
	n := reflect.New(typ).Interface()
	if err := json.Unmarshal(hit.Source, n); err != nil {
		return reflect.Value{}, err
	}
	val := reflect.ValueOf(n).Elem()
	fieldID := val.FieldByName("ID")
	if fieldID.IsValid() && fieldID.CanSet() {
		fieldID.SetString(hit.ID)
	}
	return val, nil
}
