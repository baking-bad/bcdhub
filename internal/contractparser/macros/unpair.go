package macros

import (
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
)

type unpairFamily struct{}

func (f unpairFamily) Find(arr ...*fastjson.Value) (Macros, error) {
	if len(arr) != 3 {
		return nil, nil
	}

	prim0 := getPrim(arr[0])
	prim1 := getPrim(arr[1])
	prim2 := getPrim(arr[2])
	if prim0 != pDUP || prim1 != pCAR || prim2 != pDIP {
		return nil, nil
	}

	dipArgs := getArgs(arr[2])
	if len(dipArgs) != 1 {
		return nil, nil
	}
	if dipArgs[0].Type() != fastjson.TypeArray {
		return nil, nil
	}
	primCdr := arr[2].Get("args", "0", "0", "prim").String()
	if primCdr != pCDR {
		return nil, nil
	}

	return unpairMacros{}, nil
}

type unpairMacros struct{}

func (f unpairMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in unpairMacros.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString("UNPAIR")
	newValue.Set("prim", newPrim)

	carAnnots := tree.Get("1", "annots", "0")
	if carAnnots != nil {
		cdrAnnots := tree.Get("2", "args", "0", "0", "annots", "0")
		if cdrAnnots != nil {
			newAnnots := arena.NewArray()
			newAnnots.SetArrayItem(0, carAnnots)
			newAnnots.SetArrayItem(1, cdrAnnots)
			newValue.Set("annots", newAnnots)
		}
	}

	*tree = *newValue

	return nil
}
