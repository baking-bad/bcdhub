package macros

import (
	"fmt"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type dupNMacros struct {
	*defaultMacros

	NewValue map[string]interface{}
}

func newDupNMacros() *dupNMacros {
	return &dupNMacros{
		defaultMacros: &defaultMacros{
			Name: "DUU+P",
		},
	}
}

func (m *dupNMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}

	arr := data.Array()
	if len(arr) != 2 {
		return false
	}
	firstPrim := getPrim(arr[1])
	switch firstPrim {
	case swap:
		dipPrim := getPrim(arr[0])
		if dipPrim != dip {
			return false
		}
		args := arr[0].Get("args.0.0")
		dupPrim := getPrim(args)
		if dupPrim != dup {
			return false
		}
		m.NewValue = map[string]interface{}{
			"prim": "DUUP",
		}
		return true
	case dig:
		dipPrim := getPrim(arr[0])
		if isDip(dipPrim) {
			return false
		}

		dipDepth := len(dipPrim) - 2
		depth := int(arr[1].Get("args.0.int").Int())
		if dipDepth != depth-1 {
			return false
		}
		dupPrim := arr[0].Get("args.0.prim").String()
		if dupPrim != dup {
			return false
		}
		m.NewValue = map[string]interface{}{
			"prim": fmt.Sprintf("D%sP", strings.Repeat("U", depth)),
		}
		return true
	default:
		return false
	}
}

func (m *dupNMacros) Collapse(data gjson.Result) {
	switch m.NewValue["prim"] {
	case "DUUP":
		annots := data.Get("0.args.0.0.annots")
		if annots.Exists() {
			m.NewValue["annots"] = annots.Value()
		}
	default:
		annots := data.Get("0.args.1.0.annots")
		if annots.Exists() {
			m.NewValue["annots"] = annots.Value()
		}
	}
}

func (m *dupNMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path, m.NewValue)
	m.NewValue = nil
	return
}

type dipNMacros struct {
	*defaultMacros

	NewValue map[string]interface{}
}

func newDipNMacros() *dipNMacros {
	return &dipNMacros{
		defaultMacros: &defaultMacros{
			Name: "DII+P",
		},
	}
}

func (m *dipNMacros) Find(data gjson.Result) bool {
	if !data.IsObject() {
		return false
	}
	prim := getPrim(data)
	if prim != dip {
		return false
	}
	if data.Get("args.#").Int() != 2 {
		return false
	}
	if !data.Get("args.0.int").Exists() {
		return false
	}
	return true
}

func (m *dipNMacros) Collapse(data gjson.Result) {
	depth := data.Get("args.0.int").Int()
	instr := data.Get("args.1")
	m.NewValue = map[string]interface{}{
		"prim": fmt.Sprintf("D%sP", strings.Repeat("I", int(depth))),
		"args": []interface{}{instr.Value()},
	}
}

func (m *dipNMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path, m.NewValue)
	m.NewValue = nil
	return
}
