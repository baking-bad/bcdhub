package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

//
//  STRING
//

// String -
type String struct {
	Default
}

// NewString -
func NewString(depth int) *String {
	return &String{
		Default: NewDefault(consts.STRING, 0, depth),
	}
}

// ToJSONSchema -
func (s *String) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(s.Default), nil
}

// Compare -
func (s *String) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*String)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return strings.Compare(s.Value.(string), secondItem.Value.(string)), nil
}

// Distinguish -
func (s *String) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*String)
	if !ok {
		return nil, nil
	}
	return s.Default.Distinguish(&second.Default)
}

// FindByName -
func (s *String) FindByName(name string, isEntrypoint bool) Node {
	if s.GetName() == name {
		return s
	}
	return nil
}
