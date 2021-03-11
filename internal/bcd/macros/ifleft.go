package macros

import (
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
)

type ifLeftFamily struct{}

func (f ifLeftFamily) Find(arr ...*fastjson.Value) (Macros, error) {
	if len(arr) < 1 {
		return nil, nil
	}
	tree := arr[0]
	if tree.Type() != fastjson.TypeObject {
		return nil, nil
	}
	ifLeft := getPrim(tree)
	if ifLeft != pIFLEFT {
		return nil, nil
	}

	args := getArgs(tree)
	if len(args) != 2 {
		return nil, nil
	}
	if args[0].Type() == fastjson.TypeArray && args[1].Type() == fastjson.TypeObject {
		prim := getPrim(args[1])
		renameArr, err := args[0].Array()
		if err != nil {
			return nil, err
		}
		if prim == pFAIL && len(renameArr) == 1 && getPrim(renameArr[0]) == pRENAME {
			return assertLeft{}, nil
		}
	} else if args[0].Type() == fastjson.TypeObject && args[1].Type() == fastjson.TypeArray {
		prim := getPrim(args[0])
		renameArr, err := args[1].Array()
		if err != nil {
			return nil, err
		}
		if prim == pFAIL && len(renameArr) == 1 && getPrim(renameArr[0]) == pRENAME {
			return assertRight{}, nil
		}
	}
	return nil, nil
}

type assertLeft struct{}

func (f assertLeft) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in assertLeft.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newPrim := arena.NewString("ASSERT_LEFT")
	newValue := arena.NewObject()

	newValue.Set("prim", newPrim)

	annots := tree.Get("0", "args", "0", "0", "annots")
	if annots != nil {
		newValue.Set("annots", annots)
	}

	*tree = *newValue
	return nil
}

type assertRight struct{}

func (f assertRight) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in assertRight.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newPrim := arena.NewString("ASSERT_RIGHT")
	newValue := arena.NewObject()

	newValue.Set("prim", newPrim)

	annots := tree.Get("0", "args", "1", "0", "annots")
	if annots != nil {
		newValue.Set("annots", annots)
	}

	*tree = *newValue

	return nil
}
