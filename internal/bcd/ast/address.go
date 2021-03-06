package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
)

//
//  Address
//

// Address -
type Address struct {
	Default
}

// NewAddress -
func NewAddress(depth int) *Address {
	return &Address{
		Default: NewDefault(consts.ADDRESS, 0, depth),
	}
}

// ToBaseNode -
func (a *Address) ToBaseNode(optimized bool) (*base.Node, error) {
	val := a.Value.(string)
	if a.ValueKind == valueKindBytes {
		return toBaseNodeBytes(val), nil
	}
	if optimized {
		value, err := forge.Contract(val)
		if err != nil {
			return nil, err
		}
		return toBaseNodeBytes(value), nil
	}
	return toBaseNodeString(val), nil
}

// ToMiguel -
func (a *Address) ToMiguel() (*MiguelNode, error) {
	name := a.GetTypeName()
	var value string
	if a.Value != nil {
		value = a.Value.(string)
		if a.ValueKind == valueKindBytes {
			v, err := forge.UnforgeContract(value)
			if err != nil {
				return nil, err
			}
			value = v
		}
	}
	return &MiguelNode{
		Prim:  a.Prim,
		Type:  strings.ToLower(a.Prim),
		Value: value,
		Name:  &name,
	}, nil
}

// GetJSONModel -
func (a *Address) GetJSONModel(model JSONModel) {
	value := a.Value.(string)
	if a.Value != nil {
		if a.ValueKind == valueKindBytes {
			v, err := forge.UnforgeContract(value)
			if err == nil {
				value = v
			}
		}
	}

	model[a.GetName()] = value
}

// ToJSONSchema -
func (a *Address) ToJSONSchema() (*JSONSchema, error) {
	return getAddressJSONSchema(a.Default), nil
}

// Compare -
func (a *Address) Compare(second Comparable) (int, error) {
	secondAddress, ok := second.(*Address)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	if a.Value == secondAddress.Value {
		return 0, nil
	}
	if a.ValueKind == secondAddress.ValueKind {
		return strings.Compare(a.Value.(string), secondAddress.Value.(string)), nil
	}
	return compareNotOptimizedTypes(a.Default, secondAddress.Default, forge.Contract)
}

// Distinguish -
func (a *Address) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Address)
	if !ok {
		return nil, nil
	}
	if err := a.optimizeStringValue(forge.UnforgeContract); err != nil {
		return nil, err
	}
	if err := second.optimizeStringValue(forge.UnforgeContract); err != nil {
		return nil, err
	}
	return a.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (a *Address) FromJSONSchema(data map[string]interface{}) error {
	setOptimizedJSONSchema(&a.Default, data, forge.UnforgeContract)
	return nil
}

// FindByName -
func (a *Address) FindByName(name string, isEntrypoint bool) Node {
	if a.GetName() == name {
		return a
	}
	return nil
}
