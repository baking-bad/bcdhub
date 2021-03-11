package ast

import "github.com/baking-bad/bcdhub/internal/bcd/consts"

//
//  bls12_381_fr
//

// BLS12381fr -
type BLS12381fr struct {
	Default
}

// NewBLS12381fr -
func NewBLS12381fr(depth int) *BLS12381fr {
	return &BLS12381fr{
		Default: NewDefault(consts.BLS12381FR, 0, depth),
	}
}

// ToJSONSchema -
func (b *BLS12381fr) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(b.Default), nil
}

// Distinguish -
func (b *BLS12381fr) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*BLS12381fr)
	if !ok {
		return nil, nil
	}
	return b.Default.Distinguish(&second.Default)
}

// FindByName -
func (b *BLS12381fr) FindByName(name string, isEntrypoint bool) Node {
	if b.GetName() == name {
		return b
	}
	return nil
}

//
//  bls12_381_g1
//

// BLS12381g1 -
type BLS12381g1 struct {
	Default
}

// NewBLS12381g1 -
func NewBLS12381g1(depth int) *BLS12381g1 {
	return &BLS12381g1{
		Default: NewDefault(consts.BLS12381G1, 0, depth),
	}
}

// ToJSONSchema -
func (b *BLS12381g1) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(b.Default), nil
}

// Distinguish -
func (b *BLS12381g1) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*BLS12381g1)
	if !ok {
		return nil, nil
	}
	return b.Default.Distinguish(&second.Default)
}

// FindByName -
func (b *BLS12381g1) FindByName(name string, isEntrypoint bool) Node {
	if b.GetName() == name {
		return b
	}
	return nil
}

//
//  bls12_381_g2
//

// BLS12381g2 -
type BLS12381g2 struct {
	Default
}

// NewBLS12381g2 -
func NewBLS12381g2(depth int) *BLS12381g2 {
	return &BLS12381g2{
		Default: NewDefault(consts.BLS12381G2, 0, depth),
	}
}

// ToJSONSchema -
func (b *BLS12381g2) ToJSONSchema() (*JSONSchema, error) {
	return getStringJSONSchema(b.Default), nil
}

// Distinguish -
func (b *BLS12381g2) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*BLS12381g2)
	if !ok {
		return nil, nil
	}
	return b.Default.Distinguish(&second.Default)
}

// FindByName -
func (b *BLS12381g2) FindByName(name string, isEntrypoint bool) Node {
	if b.GetName() == name {
		return b
	}
	return nil
}
