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
}

// Type -
type Type interface {
	ParseType(node *base.Node, id *int) error
	GetPrim() string
	GetEntrypoints() []string
}

// Value -
type Value interface {
	ParseValue(node *base.Node) error
	GetValue() interface{}
	ToMiguel() (*MiguelNode, error)
}

// Packer -
type Packer interface {
	Pack() ([]byte, error)
}

// Unpacker -
type Unpacker interface {
	Unpack() (string, error)
}
