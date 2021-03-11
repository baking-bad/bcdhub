package ast

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

//
//  Operation
//

// Operation -
type Operation struct {
	Default
}

// NewOperation -
func NewOperation(depth int) *Operation {
	return &Operation{
		Default: NewDefault(consts.OPERATION, 0, depth),
	}
}

// Distinguish -
func (o *Operation) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Operation)
	if !ok {
		return nil, nil
	}
	return o.Default.Distinguish(&second.Default)
}

// FindByName -
func (o *Operation) FindByName(name string, isEntrypoint bool) Node {
	if o.GetName() == name {
		return o
	}
	return nil
}
