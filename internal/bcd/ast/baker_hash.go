package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
)

//
//  BakerHash
//

// BakerHash -
type BakerHash struct {
	Default
}

// NewBakerHash -
func NewBakerHash(depth int) *BakerHash {
	return &BakerHash{
		Default: NewDefault(consts.BAKERHASH, 0, depth),
	}
}

// ToMiguel -
func (s *BakerHash) ToMiguel() (*MiguelNode, error) {
	name := s.GetTypeName()
	value := s.Value.(string)
	if s.ValueKind == valueKindBytes {
		v, err := encoding.EncodeBase58String(value, []byte(encoding.PrefixBakerHash))
		if err != nil {
			return nil, err
		}
		value = v
	}
	return &MiguelNode{
		Prim:  s.Prim,
		Type:  strings.ToLower(s.Prim),
		Value: value,
		Name:  &name,
	}, nil
}

// ToBaseNode -
func (s *BakerHash) ToBaseNode(optimized bool) (*base.Node, error) {
	val := s.Value.(string)
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
func (s *BakerHash) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(s.Default), nil
}

// Compare -
func (s *BakerHash) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*BakerHash)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return strings.Compare(s.Value.(string), secondItem.Value.(string)), nil
}

// Distinguish -
func (s *BakerHash) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*BakerHash)
	if !ok {
		return nil, nil
	}
	return s.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (s *BakerHash) FromJSONSchema(data map[string]interface{}) error {
	setOptimizedJSONSchema(&s.Default, data, forge.UnforgeBakerHash)
	return nil
}

// FindByName -
func (s *BakerHash) FindByName(name string, isEntrypoint bool) Node {
	if s.GetName() == name {
		return s
	}
	return nil
}
