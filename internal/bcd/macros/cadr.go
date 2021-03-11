package macros

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
)

type cadrFamily struct{}

func (f cadrFamily) Find(arr ...*fastjson.Value) (Macros, error) {
	if len(arr) < 2 {
		return nil, nil
	}
	name := "C"
	for i := range arr {
		if arr[i].Type() != fastjson.TypeObject {
			return nil, nil
		}
		prim := getPrim(arr[i])
		switch prim {
		case pCAR:
			name += "A"
		case pCDR:
			name += "D"
		default:
			return nil, nil
		}

	}
	name += "R"
	return cadrMacros{
		name:   name,
		length: len(arr),
	}, nil
}

type cadrMacros struct {
	name   string
	length int
}

func (f cadrMacros) Replace(tree *fastjson.Value) error {
	if tree.Type() != fastjson.TypeArray {
		return errors.Errorf("Invalid tree type in failMacros.Replace: %s", tree.Type())
	}

	arena := fastjson.Arena{}
	newValue := arena.NewObject()
	newPrim := arena.NewString(f.name)
	newValue.Set("prim", newPrim)

	idx := fmt.Sprintf("%d", f.length-1)
	annots := tree.Get(idx, "annots")
	if annots != nil {
		newValue.Set("annots", annots)
	}

	*tree = *newValue
	return nil
}
