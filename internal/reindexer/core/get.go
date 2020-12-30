package core

import (
	"reflect"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
	"github.com/restream/reindexer"
)

// Sizes -
const (
	DefaultSize = 10
)

// Count -
func (r *Reindexer) Count(query *reindexer.Query) (int64, error) {
	it := query.Exec()
	defer it.Close()

	if it.Error() != nil {
		return 0, it.Error()
	}
	count := it.TotalCount()
	return int64(count), nil
}

// GetOne -
func (r *Reindexer) GetOne(query *reindexer.Query, output interface{}) error {
	it := query.Exec()
	defer it.Close()

	if it.Error() != nil {
		return it.Error()
	}
	it.NextObj(output)
	return nil
}

// GetByID -
func (r *Reindexer) GetByID(ret models.Model) error {
	query := r.Query(ret.GetIndex()).WhereString("id", reindexer.EQ, ret.GetID())
	return r.GetOne(query, ret)
}

// GetByIDs -
func (r *Reindexer) GetByIDs(output interface{}, ids ...string) error {
	index, err := getIndex(output)
	if err != nil {
		return err
	}
	query := r.Query(index).
		WhereString("id", reindexer.EQ, ids...)

	return r.GetAllByQuery(query, output)
}

// GetAll -
func (r *Reindexer) GetAll(output interface{}) error {
	index, err := getIndex(output)
	if err != nil {
		return err
	}
	return r.GetAllByQuery(r.Query(index), output)
}

// GetByNetwork -
func (r *Reindexer) GetByNetwork(network string, output interface{}) error {
	index, err := getIndex(output)
	if err != nil {
		return err
	}
	query := r.Query(index).
		WhereString("network", reindexer.EQ, network).
		Sort("level", false)

	return r.GetAllByQuery(query, output)
}

// GetByNetworkWithSort -
func (r *Reindexer) GetByNetworkWithSort(network, sortField, sortOrder string, output interface{}) error {
	index, err := getIndex(output)
	if err != nil {
		return err
	}
	query := r.Query(index).
		WhereString("network", reindexer.EQ, network).
		Sort(sortField, sortOrder == "desc")

	return r.GetAllByQuery(query, output)
}

// GetAllByQuery -
func (r *Reindexer) GetAllByQuery(query *reindexer.Query, output interface{}) error {
	if query == nil {
		return ErrQueryPointerIsNil
	}

	it := query.Exec()
	defer it.Close()

	if it.Error() != nil {
		return it.Error()
	}

	if it.Count() == 0 {
		return nil
	}

	return parse(it, output)
}

// GetAllByQueryWithTotal -
func (r *Reindexer) GetAllByQueryWithTotal(query *reindexer.Query, output interface{}) (int, error) {
	if query == nil {
		return 0, ErrQueryPointerIsNil
	}

	it := query.ReqTotal().Exec()
	defer it.Close()

	if it.Error() != nil {
		return 0, it.Error()
	}

	if it.Count() == 0 {
		return 0, nil
	}

	return it.TotalCount(), parse(it, output)
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

func getIndex(output interface{}) (string, error) {
	typ, err := getElementType(output)
	if err != nil {
		return "", err
	}
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

func parse(it *reindexer.Iterator, output interface{}) error {
	typ, err := getElementType(output)
	if err != nil {
		return err
	}
	el := reflect.ValueOf(output).Elem()

	for it.Next() {
		obj := reflect.New(typ).Interface()
		it.NextObj(obj)
		val := reflect.ValueOf(obj).Elem()
		if el.Kind() == reflect.Slice {
			el.Set(reflect.Append(el, val))
		} else {
			el.Set(val)
		}
	}
	return nil
}
