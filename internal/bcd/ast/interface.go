package ast

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
)

// Node -
type Node interface {
	fmt.Stringer
	Type
	Value
}

// Type -
type Type interface {
	ParseType(node *base.Node, id *int) error
	GetPrim() string
	GetEntrypoints() []string
}

// Value -
type Value interface {
	forge.Unforger

	ParseValue(node *base.Node) error
	GetValue() interface{}
	ToMiguel() (*MiguelNode, error)
	Forge(optimized bool) ([]byte, error)
}
