package macros

import (
	"log"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ============================
// ========= CA[AD]+R =========
// ============================

type cadrMacros struct {
	*defaultMacros

	Prim   string
	Annots []string
}

func newCadrMacros() *cadrMacros {
	return &cadrMacros{
		defaultMacros: &defaultMacros{
			Name: "CA[AD]+R",
		},
	}
}

func (m *cadrMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}

	l := int(data.Get("#").Int())
	if l < 2 {
		return false
	}
	return len(data.Get("#(prim%\"C*R\")#").Array()) == l
}

func (m *cadrMacros) Collapse(data gjson.Result) {
	m.Prim = "C"
	m.Annots = make([]string, 0)
	log.Println(data)

	for _, item := range data.Array() {
		prim := item.Get("prim").String()
		m.Prim += string(prim[1])

		annots := item.Get("annots")
		if annots.Exists() {
			for _, ann := range annots.Array() {
				m.Annots = append(m.Annots, ann.String())
			}
		}
	}

	m.Prim += "R"
}

func (m *cadrMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path,
		map[string]interface{}{
			"prim":   m.Prim,
			"annots": m.Annots,
		},
	)
	m.Annots = nil
	m.Prim = ""
	return
}

// ===========================
// ========= SET_CAR =========
// ===========================

type setCarMacros struct {
	*defaultMacros

	NewValue map[string]interface{}
}

func newSetCarMacros() *setCarMacros {
	return &setCarMacros{
		defaultMacros: &defaultMacros{
			Name: setCar,
		},
	}
}

func (m *setCarMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}

	l := int(data.Get("#").Int())
	if l != 6 && l != 3 {
		return false
	}

	if l == 6 {
		if data.Get("0.prim").String() != dup || data.Get("1.prim").String() != car || data.Get("2.prim").String() != drop ||
			data.Get("3.prim").String() != cdr || data.Get("4.prim").String() != swap || data.Get("5.prim").String() != pair {
			return false
		}
	} else if l == 3 {
		if data.Get("0.prim").String() != cdr || data.Get("1.prim").String() != swap || data.Get("2.prim").String() != pair {
			return false
		}
	}
	return true
}

func (m *setCarMacros) Collapse(data gjson.Result) {
	m.NewValue = map[string]interface{}{
		"prim": m.Name,
	}
	if data.Get("#").Int() == 6 {
		m.NewValue["annots"] = data.Get("2.annots.1").Array()
	}
}

func (m *setCarMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path, m.NewValue)
	m.NewValue = nil
	return
}

// ===========================
// ========= SET_CDR =========
// ===========================

type setCdrMacros struct {
	*defaultMacros

	NewValue map[string]interface{}
}

func newSetCdrMacros() *setCdrMacros {
	return &setCdrMacros{
		defaultMacros: &defaultMacros{
			Name: setCdr,
		},
	}
}

func (m *setCdrMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}

	l := int(data.Get("#").Int())
	if l != 6 && l != 3 {
		return false
	}

	if l == 6 {
		if data.Get("0.prim").String() != dup || data.Get("1.prim").String() != cdr || data.Get("2.prim").String() != drop ||
			data.Get("3.prim").String() != car || data.Get("4.prim").String() != swap || data.Get("5.prim").String() != pair {
			return false
		}
	} else if l == 3 {
		if data.Get("0.prim").String() != car || data.Get("1.prim").String() != swap || data.Get("2.prim").String() != pair {
			return false
		}
	}
	return true
}

func (m *setCdrMacros) Collapse(data gjson.Result) {
	m.NewValue = map[string]interface{}{
		"prim": m.Name,
	}
	if data.Get("#").Int() == 6 {
		m.NewValue["annots"] = data.Get("2.annots.1").Array()
	}
}

func (m *setCdrMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path, m.NewValue)
	m.NewValue = nil
	return
}
