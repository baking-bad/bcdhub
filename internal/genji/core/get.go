package core

import (
	"reflect"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/genjidb/genji/document"
	"github.com/pkg/errors"
)

// Sizes -
const (
	DefaultSize = 10
)

// GetOne -
func (g *Genji) GetOne(builder *Builder, output interface{}) error {
	doc, err := g.QueryDocument(builder.End().String())
	if err != nil {
		return err
	}

	return document.StructScan(doc, output)
}

// GetByID -
func (g *Genji) GetByID(ret models.Model) error {
	builder := NewBuilder()

	builder.SelectAll(ret.GetIndex()).And(
		NewEq("id", ret.GetID()),
	)

	return g.GetOne(builder, ret)
}

// GetByIDs -
func (g *Genji) GetByIDs(output interface{}, ids ...string) error {
	builder := NewBuilder().And(
		NewIn("id", ids...),
	)

	return g.GetAllByQuery(builder, output)
}

// GetAll -
func (g *Genji) GetAll(output interface{}) error {
	return g.GetAllByQuery(NewBuilder(), output)
}

// GetByNetwork -
func (g *Genji) GetByNetwork(network string, output interface{}) error {
	builder := NewBuilder().And(
		NewEq("network", network),
	).SortAsc("level")

	return g.GetAllByQuery(builder, output)
}

// GetByNetworkWithSort -
func (g *Genji) GetByNetworkWithSort(network, sortField, sortOrder string, output interface{}) error {
	builder := NewBuilder().And(
		NewEq("network", network),
	)
	if sortOrder == "asc" {
		builder.SortAsc(sortField)
	} else {
		builder.SortDesc(sortField)
	}

	return g.GetAllByQuery(builder, output)
}

// GetAllByQuery -
func (g *Genji) GetAllByQuery(builder *Builder, output interface{}) error {
	if builder == nil {
		return ErrBuilderPointerIsNil
	}
	typ, err := getElementType(output)
	if err != nil {
		return err
	}
	index, err := getIndex(typ)
	if err != nil {
		return err
	}

	builder.SelectAll(index).End()

	res, err := g.Query(builder.String())
	if err != nil {
		return err
	}
	defer res.Close()

	el := reflect.ValueOf(output).Elem()
	res.Iterate(func(d document.Document) error {
		inter := reflect.New(typ).Interface()
		if err := document.StructScan(d, &inter); err != nil {
			return err
		}
		val := reflect.ValueOf(inter).Elem()
		if el.Kind() == reflect.Slice {
			el.Set(reflect.Append(el, val))
		} else {
			el.Set(val)
		}
		return nil
	})

	return nil
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
