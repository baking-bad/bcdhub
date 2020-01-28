package macros

import (
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type failMacros struct {
	*defaultMacros

	NewValue map[string]interface{}
}

func newFailMacros() *failMacros {
	return &failMacros{
		defaultMacros: &defaultMacros{
			Name: "FAIL",
		},
		NewValue: map[string]interface{}{
			"prim": "FAIL",
		},
	}
}

func (m *failMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}

	arr := data.Array()
	if len(arr) != 2 {
		return false
	}

	return getPrim(arr[0]) == unit && getPrim(arr[1]) == failwith
}

func (m *failMacros) Replace(json, path string) (res string, err error) {
	return sjson.Set(json, path, m.NewValue)
}
