package macros

import (
	"github.com/tidwall/gjson"
	"github.com/valyala/fastjson"
)

// HasMacros -
func HasMacros(tree gjson.Result, families *[]Family) (bool, error) {
	var p fastjson.Parser
	val, err := p.Parse(tree.String())
	if err != nil {
		return false, err
	}

	if err := collapse(val, families); err != nil {
		return false, err
	}

	return tree.String() != val.String(), nil
}
