package macros

import (
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
)

type setCdrFamily struct{}

func (f setCdrFamily) Find(arr ...*fastjson.Value) (Macros, error) {
	var offset int
	switch len(arr) {
	case 2:
	case 5:
		offset = 3

		prim0 := getPrim(arr[0])
		prim1 := getPrim(arr[1])
		prim2 := getPrim(arr[2])

		if prim0 != pDUP || prim1 != pCDR || prim2 != pDROP {
			return nil, nil
		}
	default:
		return nil, nil
	}

	fPrim := getPrim(arr[offset])
	sPrim := getPrim(arr[offset+1])

	if fPrim != pCAR || sPrim != pPAIR {
		return nil, nil
	}
	return setCdrMacros{}, nil
}

type setCdrMacros struct{}

func (f setCdrMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in setCdrMacros.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString("SET_CDR")

	newValue.Set("prim", newPrim)

	cdrPrim := tree.Get("1", "prim").String()
	if cdrPrim == pCDR {
		annots := tree.Get("1", "annots")
		if annots != nil {
			newValue.Set("annots", annots)
		}
	}

	*tree = *newValue
	return nil
}
