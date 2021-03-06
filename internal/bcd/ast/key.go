package ast

import (
	"encoding/hex"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
)

//
//  Key
//

// Key -
type Key struct {
	Default
}

// NewKey -
func NewKey(depth int) *Key {
	return &Key{
		Default: NewDefault(consts.KEY, 0, depth),
	}
}

// ToMiguel -
func (k *Key) ToMiguel() (*MiguelNode, error) {
	name := k.GetTypeName()
	value := k.Value.(string)
	if k.ValueKind == valueKindBytes {
		v, err := forge.UnforgePublicKey(value)
		if err != nil {
			return nil, err
		}
		value = v
	}
	return &MiguelNode{
		Prim:  k.Prim,
		Type:  strings.ToLower(k.Prim),
		Value: value,
		Name:  &name,
	}, nil
}

// ToBaseNode -
func (k *Key) ToBaseNode(optimized bool) (*base.Node, error) {
	val := k.Value.(string)
	if optimized {
		value, err := forge.PublicKey(val)
		if err != nil {
			return nil, err
		}
		s := hex.EncodeToString(value)
		return toBaseNodeBytes(s), nil
	}
	return toBaseNodeString(val), nil
}

// ToJSONSchema -
func (k *Key) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(k.Default), nil
}

// Compare -
func (k *Key) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Key)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return strings.Compare(k.Value.(string), secondItem.Value.(string)), nil
}

// Distinguish -
func (k *Key) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Key)
	if !ok {
		return nil, nil
	}
	return k.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (k *Key) FromJSONSchema(data map[string]interface{}) error {
	setOptimizedJSONSchema(&k.Default, data, forge.UnforgePublicKey)
	return nil
}

// FindByName -
func (k *Key) FindByName(name string, isEntrypoint bool) Node {
	if k.GetName() == name {
		return k
	}
	return nil
}
