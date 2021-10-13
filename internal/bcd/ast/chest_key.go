package ast

import (
	"encoding/hex"
	"strings"
	"unicode/utf8"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
)

// ChestKey -
type ChestKey struct {
	Default
}

// NewChestKey -
func NewChestKey(depth int) *ChestKey {
	return &ChestKey{
		Default: NewDefault(consts.CHESTKEY, 0, depth),
	}
}

// ToJSONSchema -
func (c *ChestKey) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(c.Default), nil
}

// Compare -
func (c *ChestKey) Compare(second Comparable) (int, error) {
	s, ok := second.(*ChestKey)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return strings.Compare(c.Value.(string), s.Value.(string)), nil
}

// Distinguish -
func (c *ChestKey) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*ChestKey)
	if !ok {
		return nil, nil
	}

	return c.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (c *ChestKey) FromJSONSchema(data map[string]interface{}) error {
	setBytesJSONSchema(&c.Default, data)
	return nil
}

// FindByName -
func (c *ChestKey) FindByName(name string, isEntrypoint bool) Node {
	if c.GetName() == name {
		return c
	}
	return nil
}

// ToMiguel -
func (c *ChestKey) ToMiguel() (*MiguelNode, error) {
	node, err := c.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	if str, ok := node.Value.(string); ok {
		tree := forge.TryUnpackString(str)
		if tree != nil {
			treeJSON, err := json.MarshalToString(tree)
			if err == nil {
				node.Value, _ = formatter.MichelineToMichelsonInline(treeJSON)
			}
		} else if data, err := hex.DecodeString(str); err == nil && utf8.Valid(data) {
			node.Value = string(data)
		}
	}

	return node, nil
}
