package ast

import (
	"encoding/hex"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
)

//
//  KeyHash
//

// KeyHash -
type KeyHash struct {
	Default
}

// NewKeyHash -
func NewKeyHash(depth int) *KeyHash {
	return &KeyHash{
		Default: NewDefault(consts.KEYHASH, 0, depth),
	}
}

// ToMiguel -
func (k *KeyHash) ToMiguel() (*MiguelNode, error) {
	name := k.GetTypeName()
	value := k.Value.(string)
	if k.ValueKind == valueKindBytes {
		v, err := forge.UnforgeAddress(value)
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
func (k *KeyHash) ToBaseNode(optimized bool) (*base.Node, error) {
	val := k.Value.(string)
	if optimized {
		value, err := forge.Address(val, true)
		if err != nil {
			return nil, err
		}
		s := hex.EncodeToString(value)
		return toBaseNodeBytes(s), nil
	}
	return toBaseNodeString(val), nil
}

// ToJSONSchema -
func (k *KeyHash) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(k.Default), nil
}

// Compare -
func (k *KeyHash) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*KeyHash)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return strings.Compare(k.Value.(string), secondItem.Value.(string)), nil
}

// Distinguish -
func (k *KeyHash) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*KeyHash)
	if !ok {
		return nil, nil
	}
	return k.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (k *KeyHash) FromJSONSchema(data map[string]interface{}) error {
	setOptimizedJSONSchema(&k.Default, data, forge.UnforgeAddress)
	return nil
}

// FindByName -
func (k *KeyHash) FindByName(name string, isEntrypoint bool) Node {
	if k.GetName() == name {
		return k
	}
	return nil
}
