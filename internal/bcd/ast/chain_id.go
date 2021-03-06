package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
)

//
//  ChainID
//

// ChainID -
type ChainID struct {
	Default
}

// NewChainID -
func NewChainID(depth int) *ChainID {
	return &ChainID{
		Default: NewDefault(consts.CHAINID, 0, depth),
	}
}

// ToMiguel -
func (c *ChainID) ToMiguel() (*MiguelNode, error) {
	name := c.GetTypeName()
	value := c.Value.(string)
	if c.ValueKind == valueKindBytes {
		v, err := encoding.EncodeBase58String(value, []byte(encoding.PrefixChainID))
		if err != nil {
			return nil, err
		}
		value = v
	}
	return &MiguelNode{
		Prim:  c.Prim,
		Type:  strings.ToLower(c.Prim),
		Value: value,
		Name:  &name,
	}, nil
}

// ToBaseNode -
func (c *ChainID) ToBaseNode(optimized bool) (*base.Node, error) {
	val := c.Value.(string)
	if optimized {
		value, err := encoding.DecodeBase58ToString(val)
		if err != nil {
			return nil, err
		}
		return toBaseNodeBytes(value), nil
	}
	return toBaseNodeString(val), nil
}

// ToJSONSchema -
func (c *ChainID) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(c.Default), nil
}

// Compare -
func (c *ChainID) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*ChainID)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	if c.Value == secondItem.Value {
		return 0, nil
	}
	if c.ValueKind == secondItem.ValueKind {
		return strings.Compare(c.Value.(string), secondItem.Value.(string)), nil
	}

	return compareNotOptimizedTypes(c.Default, secondItem.Default, encoding.DecodeBase58ToString)
}

// Distinguish -
func (c *ChainID) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*ChainID)
	if !ok {
		return nil, nil
	}
	return c.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (c *ChainID) FromJSONSchema(data map[string]interface{}) error {
	setOptimizedJSONSchema(&c.Default, data, forge.UnforgeChainID)
	return nil
}

// FindByName -
func (c *ChainID) FindByName(name string, isEntrypoint bool) Node {
	if c.GetName() == name {
		return c
	}
	return nil
}
