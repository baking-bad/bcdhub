package macros

import (
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
)

type ifNoneFamily struct{}

func (f ifNoneFamily) Find(arr ...*fastjson.Value) (Macros, error) {
	if len(arr) < 1 {
		return nil, nil
	}
	tree := arr[0]
	if tree.Type() != fastjson.TypeObject {
		return nil, nil
	}
	ifNone := getPrim(tree)
	if ifNone != pIFNONE {
		return nil, nil
	}

	args := getArgs(tree)
	if len(args) != 2 {
		return nil, nil
	}
	if args[0].Type() == fastjson.TypeObject && args[1].Type() == fastjson.TypeArray {
		firstPrim := getPrim(args[0])
		secondPrim := getPrim(args[1].Get("0"))
		if firstPrim == pFAIL && secondPrim == pRENAME {
			return assertSome{}, nil
		}
	} else if args[1].Type() == fastjson.TypeObject && args[0].Type() == fastjson.TypeArray {
		secondPrim := getPrim(args[1])
		args0, err := args[0].Array()
		if err != nil {
			return nil, err
		}
		if secondPrim == pFAIL && len(args0) == 0 {
			return assertNone{}, nil
		}
	}

	return nil, nil
}

type assertNone struct{}

func (f assertNone) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in assertNone.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newPrim := arena.NewString("ASSERT_NONE")
	newValue := arena.NewObject()
	newValue.Set("prim", newPrim)

	*tree = *newValue
	return nil
}

type assertSome struct{}

func (f assertSome) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in assertSome.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newPrim := arena.NewString("ASSERT_SOME")
	newValue := arena.NewObject()
	newValue.Set("prim", newPrim)

	annots := tree.Get("0", "args", "1", "0", "annots")
	if annots != nil {
		newValue.Set("annots", annots)
	}

	*tree = *newValue
	return nil
}
