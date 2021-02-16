package ast

import (
	"reflect"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

// Copy -
func Copy(node Node) Node {
	el := reflect.ValueOf(node).Elem()
	t := el.Type()

	obj := reflect.New(t)
	if obj.Kind() == reflect.Ptr {
		obj = obj.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !strings.HasSuffix(f.Name, "Type") {
			fv := obj.FieldByName(f.Name)
			if fv.CanSet() {
				val := el.FieldByName(f.Name)
				switch val.Kind() {
				case reflect.Slice:
					sl := reflect.MakeSlice(val.Type(), 0, 0)
					for i := 0; i < val.Len(); i++ {
						item := val.Index(i).Interface()
						if n, ok := item.(Node); ok {
							sl = reflect.Append(sl, reflect.ValueOf(Copy(n)))
						} else {
							sl = reflect.Append(sl, val.Index(i))
						}
					}
					fv.Set(sl)
				case reflect.Map:
					fv.Set(reflect.MakeMap(val.Type()))
				case reflect.Ptr:
					if i, ok := val.Interface().(Node); ok {
						fv.Set(reflect.ValueOf(Copy(i)))
					} else if _, ok := val.Interface().(*OrderedMap); ok {
						fv.Set(reflect.ValueOf(NewOrderedMap()))
					} else {
						fv.Set(val)
					}
				case reflect.Struct:
					s := reflect.New(val.Type()).Elem()
					for i := 0; i < val.NumField(); i++ {
						sf := s.Field(i)
						if sf.CanSet() {
							sf.Set(val.Field(i))
						}
					}
					fv.Set(s)
				case reflect.Interface:
					if i, ok := val.Interface().(Node); ok {
						fv.Set(reflect.ValueOf(Copy(i)))
					} else {
						fv.Set(val)
					}
				default:
					fv.Set(reflect.New(val.Type()).Elem())
				}
			}
			continue
		}
		val := el.FieldByName(f.Name)
		copyVal := Copy(val.Interface().(Node))
		fv := obj.FieldByName(f.Name)
		fv.Set(reflect.ValueOf(copyVal))
	}

	return obj.Addr().Interface().(Node)
}

func toBaseNodeInt(val *base.BigInt) *base.Node {
	return &base.Node{
		IntValue: val,
	}
}

func toBaseNodeString(val string) *base.Node {
	return &base.Node{
		StringValue: &val,
	}
}

func toBaseNodeBytes(val string) *base.Node {
	return &base.Node{
		BytesValue: &val,
	}
}

func mapToBaseNodes(data *OrderedMap, optimized bool) (*base.Node, error) {
	if data == nil {
		return nil, nil
	}
	node := new(base.Node)
	node.Prim = consts.PrimArray
	node.Args = make([]*base.Node, 0)

	err := data.Range(func(key, value Comparable) (bool, error) {
		keyNode, err := key.(Node).ToBaseNode(optimized)
		if err != nil {
			return true, err
		}
		var valueNode *base.Node
		if value != nil {
			valueNode, err = value.(Node).ToBaseNode(optimized)
			if err != nil {
				return true, err
			}
		}
		node.Args = append(node.Args, &base.Node{
			Prim: consts.Elt,
			Args: []*base.Node{
				keyNode, valueNode,
			},
		})
		return false, nil
	})
	return node, err
}

func arrayToBaseNode(data []Node, optimized bool) (*base.Node, error) {
	node := new(base.Node)
	node.Prim = consts.PrimArray
	node.Args = make([]*base.Node, 0)
	for i := range data {
		arg, err := data[i].ToBaseNode(optimized)
		if err != nil {
			return nil, err
		}
		node.Args = append(node.Args, arg)
	}
	return node, nil
}
