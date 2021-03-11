package macros

import (
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
)

type failFamily struct{}

func (f failFamily) Find(arr ...*fastjson.Value) (Macros, error) {
	if len(arr) < 1 {
		return nil, nil
	}
	tree := arr[0]
	if tree.Type() != fastjson.TypeArray {
		return nil, nil
	}

	failArr, err := tree.Array()
	if err != nil {
		return nil, err
	}

	if len(failArr) != 2 {
		return nil, err
	}

	unit := getPrim(failArr[0])
	failwith := getPrim(failArr[1])
	if unit == pUNIT && failwith == pFAILWITH {
		return failMacros{}, nil
	}
	return nil, nil
}

type failMacros struct{}

func (f failMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in failMacros.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString("FAIL")

	newValue.Set("prim", newPrim)
	*tree = *newValue

	return nil
}
