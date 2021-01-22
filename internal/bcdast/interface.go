package bcdast

import "fmt"

// AstNode -
type AstNode interface {
	fmt.Stringer
	AstType
	AstValue
}

// AstType -
type AstType interface {
	ParseType(untyped Untyped, id *int) error
	GetPrim() string
	GetEntrypoints() []string
}

// AstValue -
type AstValue interface {
	ParseValue(untyped Untyped) error
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
