package macros

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/valyala/fastjson"
)

var families = []family{
	failFamily{},
	ifLeftFamily{},
	ifNoneFamily{},
	unpairFamily{},
	setCarFamily{},
	setCdrFamily{},
	mapFamily{},
	ifFamily{},
}

// Collapse -
func Collapse(tree gjson.Result) (gjson.Result, error) {
	var p fastjson.Parser
	val, err := p.Parse(tree.String())
	if err != nil {
		return tree, err
	}

	if err := collapse(val); err != nil {
		return tree, err
	}

	return gjson.Parse(val.String()), nil
}

func collapse(tree *fastjson.Value) error {
	switch tree.Type() {
	case fastjson.TypeArray:
		return collapseArray(tree)
	case fastjson.TypeObject:
		return collapseObject(tree)
	default:
		return fmt.Errorf("Invalid fastjson.Type: %v", tree.Type())
	}
}

func collapseArray(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return fmt.Errorf("Invalid collapseArray fastjson.Type: %v", tree.Type())
	}

	arr, err := tree.Array()
	if err != nil {
		return err
	}

	for i := range arr {
		if err := collapse(arr[i]); err != nil {
			return err
		}
	}

	for i := range families {
		m, err := families[i].Find(arr...)
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

func collapseObject(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeObject {
		return fmt.Errorf("Invalid collapseObject fastjson.Type: %v", tree.Type())
	}

	if tree.Exists("args") {
		if err := collapseArray(tree.Get("args")); err != nil {
			return err
		}
	}

	return nil
}
