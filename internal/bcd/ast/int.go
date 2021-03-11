package ast

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

//
//  INT
//

// Int -
type Int struct {
	Default
}

// NewInt -
func NewInt(depth int) *Int {
	return &Int{
		Default: NewDefault(consts.INT, 0, depth),
	}
}

// ToJSONSchema -
func (i *Int) ToJSONSchema() (*JSONSchema, error) {
	return getIntJSONSchema(i.Default), nil
}

// Compare -
func (i *Int) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Int)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return compareBigInt(i.Default, secondItem.Default), nil
}

// Distinguish -
func (i *Int) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Int)
	if !ok {
		return nil, nil
	}
	return i.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (i *Int) FromJSONSchema(data map[string]interface{}) error {
	setIntJSONSchema(&i.Default, data)
	return nil
}

// FindByName -
func (i *Int) FindByName(name string, isEntrypoint bool) Node {
	if i.GetName() == name {
		return i
	}
	return nil
}
