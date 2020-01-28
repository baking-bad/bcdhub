package macros

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// =======================
// ======  CMPEQ  ========
// =======================

type compareMacros struct {
	*defaultMacros

	NewValues map[int]map[string]interface{}
}

func newCompareMacros() *compareMacros {
	return &compareMacros{
		defaultMacros: &defaultMacros{
			Name: "CMP",
		},
	}
}

func (m *compareMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}
	m.NewValues = make(map[int]map[string]interface{}, 0)

	arr := data.Array()
	for i, item := range arr {
		if !item.IsObject() {
			continue
		}
		prim := getPrim(item)
		if prim == compare && len(arr) > i+1 {
			eqPrim := getPrim(arr[i+1])
			if helpers.StringInArray(eqPrim, []string{
				eq, neq, lt, gt, le, ge,
			}) {
				m.NewValues[i] = nil
			}
		}
	}
	return len(m.NewValues) > 0
}

func (m *compareMacros) Collapse(data gjson.Result) {
	for current := range m.NewValues {
		res := map[string]interface{}{}
		key := fmt.Sprintf("%d", current+1)
		eqItem := data.Get(key)
		prim := getPrim(eqItem)
		res["prim"] = fmt.Sprintf("%s%s", m.Name, prim)

		annots := eqItem.Get("annots")
		if annots.Exists() {
			res["annots"] = annots.Value()
		}

		m.NewValues[current] = res
	}
}

func (m *compareMacros) Replace(json, path string) (res string, err error) {
	res = json
	for current, value := range m.NewValues {
		deleteKey := fmt.Sprintf("%s.%d", path, current+1)
		res, err = sjson.Delete(json, deleteKey)
		if err != nil {
			return
		}
		updateKey := fmt.Sprintf("%s.%d", path, current)
		res, err = sjson.Set(res, updateKey, value)
		if err != nil {
			return
		}
	}
	return
}

// =======================
// ====== IFCMPEQ ========
// =======================

type compareIfMacros struct {
	*defaultMacros

	NewValues map[int]map[string]interface{}
}

func newCompareIfMacros() *compareIfMacros {
	return &compareIfMacros{
		defaultMacros: &defaultMacros{
			Reg:  `compare,(eq|neq|lt|gt|le|ge),if\(`,
			Name: "IFCMP",
		},
	}
}

func (m *compareIfMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}
	m.NewValues = make(map[int]map[string]interface{}, 0)

	current := -1
	for i, item := range data.Array() {
		if item.IsArray() {
			continue
		}
		prim := getPrim(item)
		if prim == compare {
			current = i
		} else if current > -1 {
			if current+1 == i {
				if !helpers.StringInArray(prim, []string{
					eq, neq, lt, gt, le, ge,
				}) {
					current = -1
				}
			}
			if current+2 == i {
				if prim == ifp {
					m.NewValues[current] = nil
				} else {
					current = -1
					break
				}
			}
		}
	}
	return len(m.NewValues) > 0
}

func (m *compareIfMacros) Collapse(data gjson.Result) {
	for current := range m.NewValues {
		res := map[string]interface{}{}

		key := fmt.Sprintf("%d", current+1)
		eqItem := data.Get(key)
		prim := getPrim(eqItem)
		res["prim"] = fmt.Sprintf("%s%s", m.Name, prim)

		annots := eqItem.Get("annots")
		if annots.Exists() {
			res["annots"] = annots.Value()
		}

		key = fmt.Sprintf("%d", current+2)
		ifItem := data.Get(key)
		args := ifItem.Get("args").Array()
		if len(args) == 2 {
			res["args"] = ifItem.Get("args").Value()
		}

		m.NewValues[current] = res
	}
}

func (m *compareIfMacros) Replace(json, path string) (res string, err error) {
	res = json
	for current, value := range m.NewValues {
		deleteKey := fmt.Sprintf("%s.%d", path, current+2)
		res, err = sjson.Delete(res, deleteKey)
		if err != nil {
			return
		}
		deleteKey = fmt.Sprintf("%s.%d", path, current+1)
		res, err = sjson.Delete(res, deleteKey)
		if err != nil {
			return
		}
		updateKey := fmt.Sprintf("%s.%d", path, current)
		res, err = sjson.Set(res, updateKey, value)
		if err != nil {
			return
		}
	}
	m.NewValues = nil
	return
}

// =======================
// ======== IFEQ =========
// =======================

type ifMacros struct {
	*defaultMacros

	NewValues map[int]map[string]interface{}
}

func newIfMacros() *ifMacros {
	return &ifMacros{
		defaultMacros: &defaultMacros{
			Name: "IF",
		},
	}
}

func (m *ifMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}
	m.NewValues = make(map[int]map[string]interface{}, 0)
	arr := data.Array()

	for i, item := range arr {
		if item.IsArray() {
			continue
		}
		prim := getPrim(item)
		if prim == ifp && i >= 1 {
			eqPrim := getPrim(arr[i-1])
			if helpers.StringInArray(eqPrim, []string{
				eq, neq, lt, gt, le, ge,
			}) {
				m.NewValues[i] = nil
			}
		}
	}
	return len(m.NewValues) > 0
}

func (m *ifMacros) Collapse(data gjson.Result) {
	for current := range m.NewValues {
		res := map[string]interface{}{}

		key := fmt.Sprintf("%d", current-1)
		eqItem := data.Get(key)
		prim := getPrim(eqItem)
		res["prim"] = fmt.Sprintf("%s%s", m.Name, prim)

		annots := eqItem.Get("annots")
		if annots.Exists() {
			res["annots"] = annots.Value()
		}

		key = fmt.Sprintf("%d", current)
		ifItem := data.Get(key)
		args := ifItem.Get("args").Array()
		if len(args) == 2 {
			res["args"] = ifItem.Get("args").Value()
		}

		m.NewValues[current] = res
	}
}

func (m *ifMacros) Replace(json, path string) (res string, err error) {
	res = json
	for current, value := range m.NewValues {
		updateKey := fmt.Sprintf("%s.%d", path, current-1)
		res, err = sjson.Set(res, updateKey, value)
		if err != nil {
			return
		}
		deleteKey := fmt.Sprintf("%s.%d", path, current)
		res, err = sjson.Delete(res, deleteKey)
		if err != nil {
			return
		}
	}

	m.NewValues = nil
	return
}
