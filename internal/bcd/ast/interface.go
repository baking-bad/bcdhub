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
	GetTypeName() string
	IsPrim(prim string) bool
	IsNamed() bool
	GetEntrypoints() []string
	ToJSONSchema() (*JSONSchema, error)
	Docs(inferredName string) ([]Typedef, string, error)
	FindByName(name string) Node
}

// Value -
type Value interface {
	ParseValue(node *base.Node) error
	GetValue() interface{}
	ToMiguel() (*MiguelNode, error)
	FromJSONSchema(data map[string]interface{}) error
	EnrichBigMap(bmd []*base.BigMapDiff) error
	ToParameters() ([]byte, error)
	Equal(second Node) bool

	Comparable
	Distinguishable
}

// Base -
type Base interface {
	ToBaseNode(optimized bool) (*base.Node, error)
}

// Distinguishable -
type Distinguishable interface {
	Distinguish(second Distinguishable) (*MiguelNode, error)
}

// Comparable -
type Comparable interface {
	Compare(second Comparable) (bool, error)
}
