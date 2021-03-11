package ast

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

//
//  MUTEZ
//

// Mutez -
type Mutez struct {
	Default
}

// NewMutez -
func NewMutez(depth int) *Mutez {
	return &Mutez{
		Default: NewDefault(consts.MUTEZ, 0, depth),
	}
}

// ToJSONSchema -
func (m *Mutez) ToJSONSchema() (*JSONSchema, error) {
	return getIntJSONSchema(m.Default), nil
}

// Compare -
func (m *Mutez) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Mutez)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return compareBigInt(m.Default, secondItem.Default), nil
}

// Distinguish -
func (m *Mutez) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Mutez)
	if !ok {
		return nil, nil
	}
	return m.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (m *Mutez) FromJSONSchema(data map[string]interface{}) error {
	setIntJSONSchema(&m.Default, data)
	return nil
}

// FindByName -
func (m *Mutez) FindByName(name string, isEntrypoint bool) Node {
	if m.GetName() == name {
		return m
	}
	return nil
}
