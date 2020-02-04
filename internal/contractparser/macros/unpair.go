package macros

import (
	"fmt"
	"regexp"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type unpairMacros struct {
	*defaultMacros

	NewValue map[string]interface{}
}

func newUnpairMacros() *unpairMacros {
	return &unpairMacros{
		defaultMacros: &defaultMacros{
			Name: unpair,
		},
	}
}

func (m *unpairMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}

	if data.Get("#").Int() != 3 {
		return false
	}

	dupPrim := data.Get("0.prim").String()
	if dupPrim != dup {
		return false
	}

	carPrim := data.Get("1.prim").String()
	if carPrim != car {
		return false
	}

	return data.Get("2.args.0.0.prim").String() == cdr
}

func (m *unpairMacros) Collapse(data gjson.Result) {
	res := map[string]interface{}{
		"prim": m.Name,
	}

	carAnnots := data.Get("1.annots.0").String()
	cdrAnnots := data.Get("2.args.0.0.annots.0").String()

	annots := []string{}
	if carAnnots != "" || cdrAnnots != "" {
		annots = append(annots, carAnnots)
		annots = append(annots, cdrAnnots)
	}
	res["annots"] = annots

	m.NewValue = res
}

func (m *unpairMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path, m.NewValue)
	m.NewValue = nil
	return
}

type unpairNMacros struct {
	*defaultMacros

	Prim   string
	Annots []string
}

func newUnpairNMacros() *unpairNMacros {
	return &unpairNMacros{
		defaultMacros: &defaultMacros{
			Name: "UN[PAI]R",
		},
	}
}

func (m *unpairNMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}

	if data.Get("#").Int() < 2 {
		return false
	}

	first := data.Get("0.prim").String()
	if first != unpair {
		return false
	}

	for _, item := range data.Array() {
		prim := getPrim(item)
		isPrimDip := isDip(prim)
		if !isUnpair(prim) && !isPrimDip {
			return false
		}

		if isPrimDip {
			argPrim := item.Get("args.0.prim").String()
			if argPrim != unpair {
				return false
			}
		}
	}

	return true
}

func (m *unpairNMacros) Collapse(data gjson.Result) {
	m.Annots = make([]string, 0)
	m.Prim = "UN"
	arr := data.Array()
	for i := 0; i < len(arr); i++ {
		prim := getPrim(arr[i])
		p, annots, index := parseUnpairObject(arr, i, isDip(prim))
		m.Prim += p
		m.Annots = append(m.Annots, annots...)
		i = index
	}
	m.Prim += "R"
}

func (m *unpairNMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path, map[string]interface{}{
		"prim":   m.Prim,
		"annots": m.Annots,
	})
	m.Annots = nil
	m.Prim = ""
	return
}

func parseUnpairObject(arr []gjson.Result, idx int, deep bool) (string, []string, int) {
	if idx > len(arr)-1 {
		return "", nil, len(arr) - 1
	}

	data := arr[idx]
	if deep {
		data = data.Get("args.0")
	}

	annots := data.Get("annots").Array()
	if len(annots) == 0 {
		return "P", nil, idx
	}
	if len(annots) != 2 {
		return "", nil, idx
	}

	first, second := annots[0].String(), annots[1].String()
	if first != "" && second != "" {
		return "PAI", []string{first, second}, idx
	}
	if first == "" && second != "" {
		prim, args, index := parseUnpairObject(arr, idx+1, deep)
		prim = fmt.Sprintf("P%sI", prim)
		args = append(args, second)
		return prim, args, index
	}
	if first != "" && second == "" {
		return "PA", []string{first}, idx
	}
	return "", nil, idx
}

func isUnpair(s string) bool {
	re := regexp.MustCompile("^UNP[PAI]*R$")
	return re.MatchString(s)
}
