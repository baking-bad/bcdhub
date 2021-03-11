package macros

import (
	"github.com/valyala/fastjson"
)

// Family -
type Family interface {
	Find(...*fastjson.Value) (Macros, error)
}

// Macros -
type Macros interface {
	Replace(*fastjson.Value) error
}
