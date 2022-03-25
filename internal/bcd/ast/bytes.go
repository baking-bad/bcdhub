package ast

import (
	"encoding/hex"
	"strings"
	"unicode/utf8"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/bcd/formatter"
)

//  BYTES
//

// Bytes -
type Bytes struct {
	Default
}

// NewBytes -
func NewBytes(depth int) *Bytes {
	return &Bytes{
		Default: NewDefault(consts.BYTES, 0, depth),
	}
}

// ToJSONSchema -
func (b *Bytes) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(b.Default), nil
}

// Compare -
func (b *Bytes) Compare(second Comparable) (int, error) {
	s, ok := second.(*Bytes)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return strings.Compare(b.Value.(string), s.Value.(string)), nil
}

// Distinguish -
func (b *Bytes) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Bytes)
	if !ok {
		return nil, nil
	}

	return b.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (b *Bytes) FromJSONSchema(data map[string]interface{}) error {
	return setBytesJSONSchema(&b.Default, data)
}

// FindByName -
func (b *Bytes) FindByName(name string, isEntrypoint bool) Node {
	if b.GetName() == name {
		return b
	}
	return nil
}

// ToMiguel -
func (b *Bytes) ToMiguel() (*MiguelNode, error) {
	node, err := b.Default.ToMiguel()
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
