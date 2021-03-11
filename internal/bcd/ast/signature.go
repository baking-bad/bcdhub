package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
)

//
//  Signature
//

// Signature -
type Signature struct {
	Default
}

// NewSignature -
func NewSignature(depth int) *Signature {
	return &Signature{
		Default: NewDefault(consts.SIGNATURE, 0, depth),
	}
}

// ToMiguel -
func (s *Signature) ToMiguel() (*MiguelNode, error) {
	name := s.GetTypeName()
	value := s.Value.(string)
	if s.ValueKind == valueKindBytes {
		v, err := encoding.EncodeBase58String(value, []byte(encoding.PrefixGenericSignature))
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
func (s *Signature) ToBaseNode(optimized bool) (*base.Node, error) {
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
func (s *Signature) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(s.Default), nil
}

// Compare -
func (s *Signature) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Signature)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return strings.Compare(s.Value.(string), secondItem.Value.(string)), nil
}

// Distinguish -
func (s *Signature) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Signature)
	if !ok {
		return nil, nil
	}
	return s.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (s *Signature) FromJSONSchema(data map[string]interface{}) error {
	setOptimizedJSONSchema(&s.Default, data, forge.UnforgeSignature)
	return nil
}

// FindByName -
func (s *Signature) FindByName(name string, isEntrypoint bool) Node {
	if s.GetName() == name {
		return s
	}
	return nil
}
