package ast

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

//
//  NEVER
//

// Never -
type Never struct {
	Default
}

// NewNever -
func NewNever(depth int) *Never {
	return &Never{
		Default: NewDefault(consts.NEVER, 0, depth),
	}
}

// Compare -
func (n *Never) Compare(second Comparable) (int, error) {
	if _, ok := second.(*Never); !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return 0, nil
}

// Distinguish -
func (n *Never) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Never)
	if !ok {
		return nil, nil
	}
	return n.Default.Distinguish(&second.Default)
}

// GetJSONModel -
func (n *Never) GetJSONModel(model JSONModel) {}

// FindByName -
func (n *Never) FindByName(name string, isEntrypoint bool) Node {
	if n.GetName() == name {
		return n
	}
	return nil
}
