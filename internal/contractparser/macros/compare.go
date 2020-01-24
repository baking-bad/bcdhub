package macros

import (
	"fmt"
	"strings"

	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// =======================
// ======  CMPEQ  ========
// =======================

type compareMacros struct {
	*defaultMacros
}

func newCompareMacros() compareMacros {
	return compareMacros{
		&defaultMacros{
			Reg:  `compare,(eq)|(neq)|(lt)|(gt)|(le)|(ge)`,
			Name: "CMP",
		},
	}
}

func (m compareMacros) Collapse(data gjson.Result) map[string]interface{} {
	if data.IsArray() {
		res := map[string]interface{}{
			"prim": m.Name,
		}
		for _, item := range data.Array() {
			prim := strings.ToUpper(item.Get("prim").String())
			if helpers.StringInArray(prim, []string{
				"EQ", "NEQ", "LT", "GT", "LE", "GE",
			}) {
				res["prim"] = fmt.Sprintf("%s%s", res["prim"], prim)
				annots := item.Get("annots")
				if annots.Exists() {
					res["annots"] = annots.Value()
				}
				break
			}
		}
		return res
	}
	return data.Value().(map[string]interface{})
}

// =======================
// ====== IFCMPEQ ========
// =======================

type compareIfMacros struct {
	*defaultMacros
}

func newCompareIfMacros() compareIfMacros {
	return compareIfMacros{
		&defaultMacros{
			Reg:  `compare,(eq)|(neq)|(lt)|(gt)|(le)|(ge),if\(`,
			Name: "IFCMP",
		},
	}
}

func (m compareIfMacros) Collapse(data gjson.Result) map[string]interface{} {
	if data.IsArray() {
		res := map[string]interface{}{
			"prim": m.Name,
		}
		for _, item := range data.Array() {
			prim := strings.ToUpper(item.Get("prim").String())
			if helpers.StringInArray(prim, []string{
				"EQ", "NEQ", "LT", "GT", "LE", "GE",
			}) {
				res["prim"] = fmt.Sprintf("%s%s", res["prim"], prim)
				annots := item.Get("annots")
				if annots.Exists() {
					res["annots"] = annots.Value()
				}
			} else if prim == "IF" {
				args := item.Get("args")
				if args.Exists() {
					res["args"] = args.Value()
				}
			}
		}
		return res
	}
	return data.Value().(map[string]interface{})
}

// =======================
// ======== IFEQ =========
// =======================

type ifMacros struct {
	*defaultMacros
}

func newIfMacros() compareMacros {
	return compareMacros{
		&defaultMacros{Reg: `if,(eq)|(neq)|(lt)|(gt)|(le)|(ge)`},
	}
}

func (m ifMacros) Collapse(data gjson.Result) map[string]interface{} {
	if data.IsArray() {
		item := map[string]interface{}{
			"prim": "IF" + data.Get("1.prim").String(),
		}
		for k, v := range data.Get("2").Map() {
			if k != "prim" {
				item[k] = v.Value()
			}
		}
		return item

	}
	return data.Value().(map[string]interface{})
}
