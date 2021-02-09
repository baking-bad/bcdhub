package ast

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/pkg/errors"
)

//
//  UNIT
//

// Unit -
type Unit struct {
	Default
}

// NewUnit -
func NewUnit(depth int) *Unit {
	return &Unit{
		Default: NewDefault(consts.UNIT, 0, depth),
	}
}

// ToParameters -
func (u *Unit) ToParameters() ([]byte, error) {
	return []byte(`{"prim": "Unit"}`), nil
}

// Compare -
func (u *Unit) Compare(second Comparable) (bool, error) {
	_, ok := second.(*Unit)
	return ok, nil
}

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
func (s *String) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*String)
	if !ok {
		return false, nil
	}
	return s.Value == secondItem.Value, nil
}

//
//  INT
//

// Int -
type Int struct {
	Default
}

// NewInt -
func NewInt(depth int) *Int {
	return &Int{
		Default: NewDefault(consts.INT, 0, depth),
	}
}

// ToJSONSchema -
func (i *Int) ToJSONSchema() (*JSONSchema, error) {
	return getIntJSONSchema(i.Default), nil
}

// Compare -
func (i *Int) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*Int)
	if !ok {
		return false, nil
	}
	return compareBigInt(i.Default, secondItem.Default), nil
}

//
//  NAT
//

// Nat -
type Nat struct {
	Default
}

// NewNat -
func NewNat(depth int) *Nat {
	return &Nat{
		Default: NewDefault(consts.NAT, 0, depth),
	}
}

// ToJSONSchema -
func (n *Nat) ToJSONSchema() (*JSONSchema, error) {
	return getIntJSONSchema(n.Default), nil
}

// Compare -
func (n *Nat) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*Nat)
	if !ok {
		return false, nil
	}
	return compareBigInt(n.Default, secondItem.Default), nil
}

//
//  MUTEZ
//

// Mutez -
type Mutez struct {
	Default
}

// NewMutez -
func NewMutez(depth int) *Mutez {
	return &Mutez{
		Default: NewDefault(consts.MUTEZ, 0, depth),
	}
}

// ToJSONSchema -
func (m *Mutez) ToJSONSchema() (*JSONSchema, error) {
	return getIntJSONSchema(m.Default), nil
}

// Compare -
func (m *Mutez) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*Mutez)
	if !ok {
		return false, nil
	}
	return compareBigInt(m.Default, secondItem.Default), nil
}

//
//  BOOL
//

// Bool -
type Bool struct {
	Default
}

// NewBool -
func NewBool(depth int) *Bool {
	return &Bool{
		Default: NewDefault(consts.BOOL, 0, depth),
	}
}

// ParseValue -
func (b *Bool) ParseValue(node *base.Node) error {
	switch node.Prim {
	case consts.False:
		b.Value = false
	case consts.True:
		b.Value = true
	default:
		return consts.ErrInvalidPrim
	}
	return nil
}

// ToBaseNode -
func (b *Bool) ToBaseNode(optimized bool) (*base.Node, error) {
	val := b.Value.(bool)
	if val {
		return &base.Node{Prim: consts.True}, nil
	}
	return &base.Node{Prim: consts.False}, nil
}

// ToJSONSchema -
func (b *Bool) ToJSONSchema() (*JSONSchema, error) {
	return wrapObject(&JSONSchema{
		Prim:    b.Prim,
		Type:    JSONSchemaTypeBool,
		Default: false,
		Title:   b.GetName(),
	}), nil
}

// Compare -
func (b *Bool) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*Bool)
	if !ok {
		return false, nil
	}
	return b.Value == secondItem.Value, nil
}

//
//  Timestamp
//

// Timestamp -
type Timestamp struct {
	Default
}

// NewTimestamp -
func NewTimestamp(depth int) *Timestamp {
	return &Timestamp{
		Default: NewDefault(consts.TIMESTAMP, 0, depth),
	}
}

// ParseValue -
func (t *Timestamp) ParseValue(node *base.Node) error {
	switch {
	case node.IntValue != nil:
		i := node.IntValue.Int64()
		if 253402300799 > i { // 31 December 9999 23:59:59 Golang time restriction
			t.Value = time.Unix(i, 0).UTC()
		} else {
			t.Value = fmt.Sprintf("%d", i)
		}
	case node.StringValue != nil:
		utc, err := time.Parse(time.RFC3339, *node.StringValue)
		if err != nil {
			return err
		}
		t.Value = utc.UTC()
	}
	return nil
}

// ToBaseNode -
func (t *Timestamp) ToBaseNode(optimized bool) (*base.Node, error) {
	switch ts := t.Value.(type) {
	case time.Time:
		if optimized {
			val := base.NewBigInt(ts.UTC().Unix())
			return toBaseNodeInt(val), nil
		}
		val := ts.UTC().Format(time.RFC3339)
		return toBaseNodeBytes(val), nil
	case string:
		return toBaseNodeString(ts), nil
	}
	return nil, errors.Errorf("Invalid timestamp type")
}

// ToJSONSchema -
func (t *Timestamp) ToJSONSchema() (*JSONSchema, error) {
	return wrapObject(&JSONSchema{
		Prim:    t.Prim,
		Title:   t.GetName(),
		Type:    JSONSchemaTypeString,
		Format:  "date-time",
		Default: time.Now().UTC().Format(time.RFC3339),
	}), nil
}

// Compare -
func (t *Timestamp) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*Timestamp)
	if !ok {
		return false, nil
	}
	ts := t.Value.(time.Time)
	ts2 := secondItem.Value.(time.Time)
	return ts.Equal(ts2), nil
}

//
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
func (b *Bytes) Compare(second Comparable) (bool, error) {
	secondAddress, ok := second.(*Bool)
	if !ok {
		return false, nil
	}
	return b.Value == secondAddress.Value, nil
}

//
//  NEVER
//

// Never -
type Never struct {
	Default
}

// NewNever -
func NewNever(depth int) *Never {
	return &Never{
		Default: NewDefault(consts.NEVER, 0, depth),
	}
}

// Compare -
func (n *Never) Compare(second Comparable) (bool, error) {
	_, ok := second.(*Never)
	return ok, nil
}

//
//  Operation
//

// Operation -
type Operation struct {
	Default
}

// NewOperation -
func NewOperation(depth int) *Operation {
	return &Operation{
		Default: NewDefault(consts.OPERATION, 0, depth),
	}
}

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
func (c *ChainID) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*ChainID)
	if !ok {
		return false, nil
	}
	if c.Value == secondItem.Value {
		return true, nil
	}
	if c.valueType == secondItem.valueType {
		return false, nil
	}

	return compareNotOptimizedTypes(c.Default, secondItem.Default, encoding.DecodeBase58ToString)
}

//
//  Address
//

// Address -
type Address struct {
	Default
}

// NewAddress -
func NewAddress(depth int) *Address {
	return &Address{
		Default: NewDefault(consts.ADDRESS, 0, depth),
	}
}

// ToBaseNode -
func (a *Address) ToBaseNode(optimized bool) (*base.Node, error) {
	val := a.Value.(string)
	if a.valueType == valueTypeBytes {
		return toBaseNodeBytes(val), nil
	}
	if optimized {
		value, err := getOptimizedContract(val)
		if err != nil {
			return nil, err
		}
		return toBaseNodeBytes(value), nil
	}
	return toBaseNodeString(val), nil
}

// ToJSONSchema -
func (a *Address) ToJSONSchema() (*JSONSchema, error) {
	return getAddressJSONSchema(a.Default), nil
}

// Compare -
func (a *Address) Compare(second Comparable) (bool, error) {
	secondAddress, ok := second.(*Address)
	if !ok {
		return false, nil
	}
	if a.Value == secondAddress.Value {
		return true, nil
	}
	if a.valueType == secondAddress.valueType {
		return false, nil
	}
	return compareNotOptimizedTypes(a.Default, secondAddress.Default, getOptimizedContract)
}

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

// ToBaseNode -
func (k *Key) ToBaseNode(optimized bool) (*base.Node, error) {
	val := k.Value.(string)
	if optimized {
		value, err := getOptimizedPublicKey(val)
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
func (k *Key) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*Key)
	if !ok {
		return false, nil
	}
	return k.Value == secondItem.Value, nil
}

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

// ToBaseNode -
func (k *KeyHash) ToBaseNode(optimized bool) (*base.Node, error) {
	val := k.Value.(string)
	if optimized {
		value, err := getOptimizedAddress(val, true)
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
func (k *KeyHash) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*KeyHash)
	if !ok {
		return false, nil
	}
	return k.Value == secondItem.Value, nil
}

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
func (s *Signature) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*Signature)
	if !ok {
		return false, nil
	}
	return s.Value == secondItem.Value, nil
}

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

func compareNotOptimizedTypes(x, y Default, optimizer func(string) (string, error)) (bool, error) {
	var a Default
	var b Default
	if x.valueType == valueTypeBytes {
		a = y
		b = x
	} else {
		a = x
		b = y
	}
	value, err := optimizer(a.Value.(string))
	if err != nil {
		return false, err
	}
	return value == b.Value.(string), nil
}

func compareBigInt(x, y Default) bool {
	xi := x.Value.(*base.BigInt)
	yi := y.Value.(*base.BigInt)
	return xi.Cmp(yi.Int) == 0
}
