package ast

import (
	"encoding/hex"
	"strings"
	"unicode/utf8"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
)

// Chest -
type Chest struct {
	Default
}

// NewChest -
func NewChest(depth int) *Chest {
	return &Chest{
		Default: NewDefault(consts.CHEST, 0, depth),
	}
}

// ToJSONSchema -
func (c *Chest) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(c.Default), nil
}

// Compare -
func (c *Chest) Compare(second Comparable) (int, error) {
	s, ok := second.(*Chest)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return strings.Compare(c.Value.(string), s.Value.(string)), nil
}

// Distinguish -
func (c *Chest) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Chest)
	if !ok {
		return nil, nil
	}

	return c.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (c *Chest) FromJSONSchema(data map[string]interface{}) error {
	setBytesJSONSchema(&c.Default, data)
	return nil
}

// FindByName -
func (c *Chest) FindByName(name string, isEntrypoint bool) Node {
	if c.GetName() == name {
		return c
	}
	return nil
}

// ToMiguel -
func (c *Chest) ToMiguel() (*MiguelNode, error) {
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
