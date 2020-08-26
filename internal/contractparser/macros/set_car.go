package macros

import (
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
)

type setCarFamily struct{}

func (f setCarFamily) Find(arr ...*fastjson.Value) (Macros, error) {
	var offset int
	switch len(arr) {
	case 3:
	case 6:
		offset = 3
		prim0 := getPrim(arr[0])
		prim1 := getPrim(arr[1])
		prim2 := getPrim(arr[2])
		if prim0 != pDUP || prim1 != pCAR || prim2 != pDROP {
			return nil, nil
		}
	default:
		return nil, nil
	}

	fPrim := getPrim(arr[offset])
	sPrim := getPrim(arr[offset+1])
	tPrim := getPrim(arr[offset+2])

	if fPrim != pCDR || sPrim != pSWAP || tPrim != pPAIR {
		return nil, nil
	}
	return setCarMacros{}, nil
}

type setCarMacros struct{}

func (f setCarMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in setCarMacros.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString("SET_CAR")

	newValue.Set("prim", newPrim)

	carPrim := tree.Get("1", "prim").String()
	if carPrim == pCAR {
		annots := tree.Get("1", "annots")
		if annots != nil {
			newValue.Set("annots", annots)
		}
	}

	*tree = *newValue
	return nil
}
