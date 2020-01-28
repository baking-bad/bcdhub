package macros

import (
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type pairMacros struct {
	*defaultMacros

	Prim   string
	Args   []string
	Annots []string
}

func newPairMacros() *pairMacros {
	return &pairMacros{
		defaultMacros: &defaultMacros{
			Name: "[PAI]R",
		},
		Prim: "",
	}
}

func (m *pairMacros) Find(data gjson.Result) bool {
	if !data.IsArray() {
		return false
	}

	arr := reverseArray(data.Array())
	if len(arr) < 2 {
		return false
	}

	last := arr[0].Get("prim").String()
	if last != pair {
		return false
	}

	for _, item := range arr[1:] {
		if !item.IsObject() {
			return false
		}

		prim := getPrim(item)

		isDipPrim := isDip(prim)
		isPairPrim := isPai(prim)

		if !(isPairPrim || isDipPrim) {
			return false
		}

		if isDipPrim {
			argPrim := getPrim(item.Get("args.0.0"))
			if !isPai(argPrim) {
				return false
			}
		}
	}
	return true
}

func (m *pairMacros) Collapse(data gjson.Result) {
	m.Args = make([]string, 0)
	m.Annots = make([]string, 0)

	arr := data.Get("@reverse").Array()
	for i := 0; i < len(arr); i++ {
		prim := getPrim(arr[i])

		p, args, annots, index := parsePairObject(arr, i, isDip(prim))
		m.Prim += p
		m.Args = append(m.Args, args...)
		m.Annots = append(m.Annots, annots...)
		i = index

	}
	m.Prim += "R"
}

func (m *pairMacros) Replace(json, path string) (res string, err error) {
	res, err = sjson.Set(json, path,
		map[string]interface{}{
			"prim":   m.Prim,
			"args":   m.Args,
			"annots": m.Annots,
		},
	)
	m.Args = nil
	m.Annots = nil
	m.Prim = ""
	return
}

func parsePairObject(arr []gjson.Result, idx int, deep bool) (string, []string, []string, int) {
	if idx > len(arr)-1 {
		return "", nil, nil, len(arr)
	}

	data := arr[idx]
	if deep {
		data = data.Get("args.0.0")
	}
	annots := data.Get("annots")
	if !annots.Exists() || len(annots.Array()) == 0 {
		return "P", nil, nil, idx
	}

	annotsArr := make([]string, 0)
	for _, ann := range annots.Array() {
		annotsArr = append(annotsArr, ann.String())
	}

	switch len(annotsArr) {
	case 1:
		annType := parseAnnotType(annotsArr[0])
		if annType == 1 {
			return "P", nil, annotsArr, idx
		}
		if annType == 2 {
			return "PA", annotsArr, nil, idx
		}
	case 2:
		annType := parseAnnotType(annotsArr[1])
		if annType == 1 {
			return "PA", annotsArr[0:1], annotsArr[1:2], idx
		}
		if annType == 2 {
			annType0 := parseAnnotType(annotsArr[0])
			if annType0 == 3 {
				prim, a, ann, index := parsePairObject(arr, idx+1, deep)
				retPrim := fmt.Sprintf("P%sI", prim)
				a = append(a, annotsArr[1])
				return retPrim, a, ann, index
			}
			return "PAI", annotsArr, nil, idx
		}
	case 3:
		return "PAI", annotsArr[:2], annotsArr[2:], idx
	default:
		return "", nil, nil, idx
	}
	return "", nil, nil, idx
}

func parseAnnotType(s string) int {
	if len(s) == 0 {
		return 0
	}
	if s[0] == '@' {
		return 1
	}
	if s[0] == '%' {
		if len(s) > 1 {
			return 2
		}
		return 3
	}
	return 0
}

func reverseArray(s []gjson.Result) []gjson.Result {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}
