package macros

import (
	"github.com/valyala/fastjson"
)

type family interface {
	Find(...*fastjson.Value) (macros, error)
}

type macros interface {
	Replace(*fastjson.Value, int) error
	Skip() int
}
