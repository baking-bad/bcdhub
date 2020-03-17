package macros

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// =======================
// ======  CMPEQ  ========
// =======================

type compareMacros struct {
	*defaultMacros

	NewValues []map[string]interface{}
	Indices   []int
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
	m.Indices = make([]int, 0)

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
				m.Indices = append(m.Indices, i)
			}
		}
	}
	return len(m.NewValues) > 0
}

func (m *compareMacros) Collapse(data gjson.Result) {
	m.NewValues = make([]map[string]interface{}, 0)

	for i, current := range m.Indices {
		res := map[string]interface{}{}

		key := fmt.Sprintf("%d", current+1)
		eqItem := data.Get(key)
		prim := getPrim(eqItem)
		res["prim"] = fmt.Sprintf("%s%s", m.Name, prim)

		if annots := eqItem.Get("annots"); annots.Exists() {
			m.NewValues[i]["annots"] = annots.Value()
		}

		m.NewValues[i] = res
	}
}

func (m *compareMacros) Replace(json, path string) (res string, err error) {
	res = json
	for i, current := range m.Indices {
		updateKey := fmt.Sprintf("%s.%d", path, current-i)
		res, err = sjson.Set(res, updateKey, m.NewValues[i])
		if err != nil {
			return
		}
		deleteKey := fmt.Sprintf("%s.%d", path, current+1-i)
		res, err = sjson.Delete(res, deleteKey)
		if err != nil {
			return
		}
	}
	m.Indices = nil
	m.NewValues = nil
	return
}

// =======================
// ====== IFCMPEQ ========
// =======================

type compareIfMacros struct {
	*defaultMacros

	NewValues []map[string]interface{}
	Indices   []int
}

func newCompareIfMacros() *compareIfMacros {
	return &compareIfMacros{
		defaultMacros: &defaultMacros{
			Name: ifcmp,
		},
	}
}

func (m *compareIfMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}
	m.Indices = make([]int, 0)

	arr := data.Array()
	for i, item := range arr {
		if item.IsArray() {
			continue
		}
		prim := getPrim(item)
		if prim == compare && len(arr) > i+2 {
			firstPrim := getPrim(arr[i+1])
			if !helpers.StringInArray(firstPrim, []string{
				eq, neq, lt, gt, le, ge,
			}) {
				continue
			}
			secondPrim := getPrim(arr[i+2])
			if secondPrim == ifp {
				m.Indices = append(m.Indices, i)
			}
		}
	}
	return len(m.NewValues) > 0
}

func (m *compareIfMacros) Collapse(data gjson.Result) {
	m.NewValues = make([]map[string]interface{}, len(m.Indices))
	for i, current := range m.Indices {
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
		if args := ifItem.Get("args"); args.Exists() && len(args.Array()) == 2 {
			res["args"] = args.Value()
		}

		m.NewValues[i] = res
	}
}

func (m *compareIfMacros) Replace(json, path string) (res string, err error) {
	res = json
	for i, current := range m.Indices {
		updateKey := fmt.Sprintf("%s.%d", path, current-2*i)
		res, err = sjson.Set(res, updateKey, m.NewValues[i])
		if err != nil {
			return
		}
		deleteKey := fmt.Sprintf("%s.%d", path, current+1-2*i)
		res, err = sjson.Delete(res, deleteKey)
		if err != nil {
			return
		}
		deleteKey = fmt.Sprintf("%s.%d", path, current+2-2*i)
		res, err = sjson.Delete(res, deleteKey)
		if err != nil {
			return
		}
	}
	m.NewValues = nil
	m.Indices = nil
	return
}

// =======================
// ======== IFEQ =========
// =======================

type ifMacros struct {
	*defaultMacros

	NewValues []map[string]interface{}
	Indices   []int
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
	m.Indices = make([]int, 0)

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
				m.Indices = append(m.Indices, i-1)
			}
		}
	}
	return len(m.NewValues) > 0
}

func (m *ifMacros) Collapse(data gjson.Result) {
	m.NewValues = make([]map[string]interface{}, len(m.Indices))
	for i, current := range m.Indices {
		res := map[string]interface{}{}

		key := fmt.Sprintf("%d", current)
		eqItem := data.Get(key)
		prim := getPrim(eqItem)
		res["prim"] = fmt.Sprintf("%s%s", m.Name, prim)

		annots := eqItem.Get("annots")
		if annots.Exists() {
			res["annots"] = annots.Value()
		}

		key = fmt.Sprintf("%d", current-1)
		ifItem := data.Get(key)
		if args := ifItem.Get("args"); args.Exists() && len(args.Array()) == 2 {
			res["args"] = args.Value()
		}

		m.NewValues[i] = res
	}
}

func (m *ifMacros) Replace(json, path string) (res string, err error) {
	res = json

	for i, current := range m.Indices {
		updateKey := fmt.Sprintf("%s.%d", path, current-i)
		res, err = sjson.Set(res, updateKey, m.NewValues[i])
		if err != nil {
			return
		}
		deleteKey := fmt.Sprintf("%s.%d", path, current+1-i)
		res, err = sjson.Delete(res, deleteKey)
		if err != nil {
			return
		}
	}

	m.NewValues = nil
	m.Indices = nil
	return
}
