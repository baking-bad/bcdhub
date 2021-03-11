package ast

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

//
//  NAT
//

// Nat -
type Nat struct {
	Default
}

// NewNat -
func NewNat(depth int) *Nat {
	return &Nat{
		Default: NewDefault(consts.NAT, 0, depth),
	}
}

// ToJSONSchema -
func (n *Nat) ToJSONSchema() (*JSONSchema, error) {
	return getIntJSONSchema(n.Default), nil
}

// Compare -
func (n *Nat) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Nat)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return compareBigInt(n.Default, secondItem.Default), nil
}

// Distinguish -
func (n *Nat) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Nat)
	if !ok {
		return nil, nil
	}
	return n.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (n *Nat) FromJSONSchema(data map[string]interface{}) error {
	setIntJSONSchema(&n.Default, data)
	return nil
}

// FindByName -
func (n *Nat) FindByName(name string, isEntrypoint bool) Node {
	if n.GetName() == name {
		return n
	}
	return nil
}
