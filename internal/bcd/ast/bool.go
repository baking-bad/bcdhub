package ast

import (
	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

//
//  BOOL
//

// Bool -
type Bool struct {
	Default
}

// NewBool -
func NewBool(depth int) *Bool {
	return &Bool{
		Default: NewDefault(consts.BOOL, 0, depth),
	}
}

// ParseValue -
func (b *Bool) ParseValue(node *base.Node) error {
	switch node.Prim {
	case consts.False:
		b.Value = false
	case consts.True:
		b.Value = true
	default:
		return consts.ErrInvalidPrim
	}
	return nil
}

// ToBaseNode -
func (b *Bool) ToBaseNode(optimized bool) (*base.Node, error) {
	val := b.Value.(bool)
	if val {
		return &base.Node{Prim: consts.True}, nil
	}
	return &base.Node{Prim: consts.False}, nil
}

// ToJSONSchema -
func (b *Bool) ToJSONSchema() (*JSONSchema, error) {
	return wrapObject(&JSONSchema{
		Prim:    b.Prim,
		Type:    JSONSchemaTypeBool,
		Default: false,
		Title:   b.GetName(),
	}), nil
}

// Compare -
func (b *Bool) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Bool)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	switch {
	case b.Value == secondItem.Value:
		return 0, nil
	case b.Value:
		return 1, nil
	default:
		return -1, nil
	}
}

// Distinguish -
func (b *Bool) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Bool)
	if !ok {
		return nil, nil
	}
	return b.Default.Distinguish(&second.Default)
}

// ToParameters -
func (b *Bool) ToParameters() ([]byte, error) {
	if v, ok := b.Value.(bool); ok && v {
		return []byte(`{"prim":"True"}`), nil
	}
	return []byte(`{"prim":"False"}`), nil
}

// FindByName -
func (b *Bool) FindByName(name string, isEntrypoint bool) Node {
	if b.GetName() == name {
		return b
	}
	return nil
}
