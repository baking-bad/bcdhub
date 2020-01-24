package macros

import "github.com/tidwall/gjson"

type failMacros struct {
	*defaultMacros
}

func newFailMacros() failMacros {
	return failMacros{
		&defaultMacros{Reg: `^\(unit,failwith\)$`},
	}
}

func (m failMacros) Collapse(data gjson.Result) map[string]interface{} {
	return map[string]interface{}{
		"prim": "FAIL",
	}
}
