package ast

import (
	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

//
//  UNIT
//

// Unit -
type Unit struct {
	Default
}

// NewUnit -
func NewUnit(depth int) *Unit {
	return &Unit{
		Default: NewDefault(consts.UNIT, 0, depth),
	}
}

// ToBaseNode -
func (u *Unit) ToBaseNode(optimized bool) (*base.Node, error) {
	return &base.Node{
		Prim: consts.Unit,
	}, nil
}

// ToParameters -
func (u *Unit) ToParameters() ([]byte, error) {
	return []byte(`{"prim":"Unit"}`), nil
}

// Compare -
func (u *Unit) Compare(second Comparable) (int, error) {
	if _, ok := second.(*Unit); !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return 0, nil
}

// Distinguish -
func (u *Unit) Distinguish(x Distinguishable) (*MiguelNode, error) {
	s, ok := x.(*Unit)
	if !ok {
		return nil, nil
	}
	return s.Default.Distinguish(&s.Default)
}

// GetJSONModel -
func (u *Unit) GetJSONModel(model JSONModel) {}

// FindByName -
func (u *Unit) FindByName(name string, isEntrypoint bool) Node {
	if u.GetName() == name {
		return u
	}
	return nil
}
