package macros

import (
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
)

type mapFamily struct{}

func (f mapFamily) Find(arr ...*fastjson.Value) (Macros, error) {
	if len(arr) < 5 {
		return nil, nil
	}

	prim1 := getPrim(arr[0])
	prim2 := getPrim(arr[1])
	if prim1 != pDUP || prim2 != pCDR {
		return nil, nil
	}

	prim3 := getPrim(arr[2])
	prim4 := getPrim(arr[3])
	prim5 := getPrim(arr[4])

	switch len(arr) {
	case 5:
		if prim3 != pDIP || prim4 != pSWAP || prim5 != pPAIR {
			return nil, nil
		}

		dipArgs := getArgs(arr[2])
		if len(dipArgs) != 1 {
			return nil, nil
		}

		seq, err := dipArgs[0].Array()
		if err != nil {
			return nil, err
		}
		if len(seq) != 2 {
			return nil, nil
		}

		dipArg0Prim := getPrim(seq[0])
		if dipArg0Prim != pCAR {
			return nil, nil
		}
		return mapCarMacros{}, nil
	case 6:
		prim6 := getPrim(arr[5])
		if prim4 != pSWAP || prim5 != pCAR || prim6 != pPAIR {
			return nil, nil
		}
		return mapCdrMacros{}, nil
	default:
		return nil, nil
	}
}

type mapCarMacros struct{}

func (f mapCarMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in mapCarMacros.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newCode := arena.NewArray()
	newPrim := arena.NewString("MAP_CAR")
	newValue.Set("prim", newPrim)

	code := tree.Get("2", "args", "0", "1")
	newCode.SetArrayItem(0, code)
	newValue.Set("args", newCode)

	annots := tree.GetArray("4", "annots")
	if len(annots) == 2 {
		newAnnots := arena.NewArray()
		newAnnots.SetArrayItem(0, annots[0])
		newValue.Set("annots", newAnnots)
	}

	*tree = *newValue
	return nil
}

type mapCdrMacros struct{}

func (f mapCdrMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in mapCdrMacros.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString("MAP_CDR")
	newCode := arena.NewArray()
	newValue.Set("prim", newPrim)

	code := tree.Get("2")
	newCode.SetArrayItem(0, code)
	newValue.Set("args", newCode)

	annots := tree.GetArray("5", "annots")
	if len(annots) == 2 {
		newAnnots := arena.NewArray()
		newAnnots.SetArrayItem(0, annots[1])
		newValue.Set("annots", newAnnots)
	}

	*tree = *newValue
	return nil
}
