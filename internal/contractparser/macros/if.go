package macros

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
)

type ifFamily struct{}

func (f ifFamily) Find(arr ...*fastjson.Value) (Macros, error) {
	switch len(arr) {
	case 1:
		if f.isAssert(arr[0]) {
			return assertMacros{}, nil
		}
	case 2:
		return f.getComplexMacros(arr...)
	default:
		return nil, nil
	}
	return nil, nil
}

func (f ifFamily) getComplexMacros(arr ...*fastjson.Value) (Macros, error) {
	if arr[0].Type() == fastjson.TypeObject {
		if isEq(getPrim(arr[0])) && getPrim(arr[1]) == pIF {
			if f.isAssert(arr[1]) {
				return assertEqMacros{}, nil
			}
			return ifEqMacros{}, nil
		}
		if isEq(getPrim(arr[1])) && getPrim(arr[0]) == pCOMPARE {
			return cmpEqMacros{}, nil
		}
		if isCmpEq(getPrim(arr[0])) && getPrim(arr[1]) == pIF {
			if f.isAssert(arr[1]) {
				return assertEqMacros{}, nil
			}
			return ifCmpEqMacros{}, nil
		}
	} else if arr[0].Type() == fastjson.TypeArray {
		flag, err := f.isCmpEqJSON(arr...)
		if err != nil {
			return nil, err
		}
		if flag {
			if f.isAssert(arr[1]) {
				return assertCmpEqMacros{}, nil
			}
			return ifCmpEqMacros{}, nil
		}
	}
	return nil, nil
}

func (f ifFamily) isCmpEqJSON(arr ...*fastjson.Value) (bool, error) {
	cmpArr, err := arr[0].Array()
	if err != nil {
		return false, err
	}
	if len(cmpArr) != 2 {
		return false, nil
	}
	if !isEq(getPrim(cmpArr[1])) {
		return false, nil
	}
	return getPrim(cmpArr[0]) == pCOMPARE && getPrim(arr[1]) == pIF, nil
}

func (f ifFamily) isAssert(tree *fastjson.Value) bool {
	if getPrim(tree) != pIF {
		return false
	}
	args := getArgs(tree)
	if len(args) != 2 {
		return false
	}

	if args[0].Type() != fastjson.TypeArray || args[1].Type() != fastjson.TypeObject {
		return false
	}
	args0 := tree.GetArray("args", "0")
	args1 := tree.GetObject("args", "1")

	return len(args0) == 0 && args1.Get("prim").String() == pFAIL
}

type assertMacros struct{}

func (f assertMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in assertMacros.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString("ASSERT")
	newValue.Set("prim", newPrim)

	*tree = *newValue

	return nil
}

type assertEqMacros struct{}

func (f assertEqMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in assertEqMacros.Replace: %s", tree.Type())
	}
	eqType := tree.GetStringBytes("0", "prim")

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString(fmt.Sprintf("ASSERT_%s", string(eqType)))
	newValue.Set("prim", newPrim)

	*tree = *newValue

	return nil
}

type assertCmpEqMacros struct{}

func (f assertCmpEqMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in assertCmpEqMacros.Replace: %s", tree.Type())
	}
	eqType := tree.GetStringBytes("0", "prim")

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString(fmt.Sprintf("ASSERT_%s", string(eqType)))
	newValue.Set("prim", newPrim)

	*tree = *newValue

	return nil
}

type cmpEqMacros struct{}

func (f cmpEqMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in cmpEqMacros.Replace: %s", tree.Type())
	}
	eqType := tree.GetStringBytes("1", "prim")

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString(fmt.Sprintf("CMP%s", eqType))
	newValue.Set("prim", newPrim)

	*tree = *newValue

	return nil
}

type ifEqMacros struct{}

func (f ifEqMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in assertMacros.Replace: %s", tree.Type())
	}
	eqType := tree.GetStringBytes("0", "prim")
	args := tree.Get("1", "args")
	if args == nil {
		return nil
	}

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString(fmt.Sprintf("IF%s", eqType))
	newValue.Set("prim", newPrim)
	newValue.Set("args", args)

	*tree = *newValue

	return nil
}

type ifCmpEqMacros struct{}

func (f ifCmpEqMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in assertMacros.Replace: %s", tree.Type())
	}
	eqType := tree.GetStringBytes("0", "prim")
	args := tree.Get("1", "args")

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString(fmt.Sprintf("IF%s", string(eqType)))
	newValue.Set("prim", newPrim)
	newValue.Set("args", args)

	*tree = *newValue

	return nil
}
