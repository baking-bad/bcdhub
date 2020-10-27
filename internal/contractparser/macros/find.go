package macros

import (
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/valyala/fastjson"
)

// Collapse -
//nolint
func Collapse(tree gjson.Result, families *[]Family) (gjson.Result, error) {
	var p fastjson.Parser
	val, err := p.Parse(tree.String())
	if err != nil {
		return tree, err
	}

	if err := collapse(val, families); err != nil {
		return tree, err
	}

	return gjson.Parse(val.String()), nil
}

func collapse(tree *fastjson.Value, families *[]Family) error {
	switch tree.Type() {
	case fastjson.TypeArray:
		return collapseArray(tree, families)
	case fastjson.TypeObject:
		return collapseObject(tree, families)
	default:
		return errors.Errorf("Invalid fastjson.Type: %v", tree.Type())
	}
}

func collapseArray(tree *fastjson.Value, families *[]Family) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid collapseArray fastjson.Type: %v", tree.Type())
	}

	arr, err := tree.Array()
	if err != nil {
		return err
	}

	for i := range arr {
		if err := collapse(arr[i], families); err != nil {
			return err
		}
	}

	for i := range *families {
		m, err := (*families)[i].Find(arr...)
		if err != nil {
			return err
		}
		if m == nil {
			continue
		}
		// log.Printf("found: %T", m)
		if err := m.Replace(tree); err != nil {
			return err
		}
		break
	}

	return nil
}

func collapseObject(tree *fastjson.Value, families *[]Family) error {
	if tree.Type() != fastjson.TypeObject {
		return errors.Errorf("Invalid collapseObject fastjson.Type: %v", tree.Type())
	}

	if tree.Exists("args") {
		if err := collapseArray(tree.Get("args"), families); err != nil {
			return err
		}
	}

	return nil
}
