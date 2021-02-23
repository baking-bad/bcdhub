package ast

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
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
	Docs(inferredName string) ([]Typedef, string, error)
	EqualType(node Node) bool
	FindByName(name string) Node
	GetEntrypoints() []string
	GetJSONModel(JSONModel)
	GetName() string
	GetPrim() string
	GetTypeName() string
	IsNamed() bool
	IsPrim(prim string) bool
	ParseType(node *base.Node, id *int) error
	Range(handler func(node Node) error) error
	ToJSONSchema() (*JSONSchema, error)
}

// Value -
type Value interface {
	Comparable
	Distinguishable

	EnrichBigMap(bmd []*types.BigMapDiff) error
	Equal(second Node) bool
	FindPointers() map[int64]*BigMap
	FromJSONSchema(data map[string]interface{}) error
	GetValue() interface{}
	ParseValue(node *base.Node) error
	ToMiguel() (*MiguelNode, error)
	ToParameters() ([]byte, error)
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
	Compare(second Comparable) (int, error)
}
