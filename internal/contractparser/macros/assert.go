package macros

import "github.com/tidwall/gjson"

// =======================
// ======= ASSERT ========
// =======================

type assertMacros struct {
	*defaultMacros
}

func newAssertMacros() assertMacros {
	return assertMacros{
		&defaultMacros{Reg: `^\(compare,(eq)|(neq)|(lt)|(gt)|(le)|(ge),.*\)`},
	}
}

func (m assertMacros) Collapse(data gjson.Result) map[string]interface{} {
	return map[string]interface{}{
		"prim": "ASSERT",
	}
}

// =======================
// ==== ASSERT_NONE ======
// =======================

type assertNoneMacros struct {
	*defaultMacros
}

func newAssertNoneMacros() assertMacros {
	return assertMacros{
		&defaultMacros{Reg: `^\(if_none\(\(\(unit,failwith\)\),\(\)\)\)$`},
	}
}

func (m assertNoneMacros) Collapse(data gjson.Result) map[string]interface{} {
	return map[string]interface{}{
		"prim": "ASSERT_NONE",
	}
}
