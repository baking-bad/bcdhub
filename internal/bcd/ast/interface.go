package ast

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
)

// Node -
type Node interface {
	fmt.Stringer
	Type
	Value
	Base
}

// Type -
type Type interface {
	ParseType(node *base.Node, id *int) error
	GetPrim() string
	GetName() string
	GetEntrypoints() []string
	ToJSONSchema() (*JSONSchema, error)
}

// Value -
type Value interface {
	ParseValue(node *base.Node) error
	GetValue() interface{}
	ToMiguel() (*MiguelNode, error)
}

// Base -
type Base interface {
	ToBaseNode(optimized bool) (*base.Node, error)
}
