package ast

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
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

// ToBaseNode -
func (u *Unit) ToBaseNode(optimized bool) (*base.Node, error) {
	return &base.Node{
		Prim: consts.Unit,
	}, nil
}

// ToParameters -
func (u *Unit) ToParameters() ([]byte, error) {
	return []byte(`{"prim":"Unit"}`), nil
}

// Compare -
func (u *Unit) Compare(second Comparable) (int, error) {
	if _, ok := second.(*Unit); !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return 0, nil
}

// Distinguish -
func (u *Unit) Distinguish(x Distinguishable) (*MiguelNode, error) {
	s, ok := x.(*Unit)
	if !ok {
		return nil, nil
	}
	return s.Default.Distinguish(&s.Default)
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
func (i *Int) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Int)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return compareBigInt(i.Default, secondItem.Default), nil
}

// Distinguish -
func (i *Int) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Int)
	if !ok {
		return nil, nil
	}
	return i.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (i *Int) FromJSONSchema(data map[string]interface{}) error {
	setIntJSONSchema(&i.Default, data)
	return nil
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
func (n *Nat) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Nat)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return compareBigInt(n.Default, secondItem.Default), nil
}

// Distinguish -
func (n *Nat) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Nat)
	if !ok {
		return nil, nil
	}
	return n.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (n *Nat) FromJSONSchema(data map[string]interface{}) error {
	setIntJSONSchema(&n.Default, data)
	return nil
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
func (m *Mutez) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Mutez)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return compareBigInt(m.Default, secondItem.Default), nil
}

// Distinguish -
func (m *Mutez) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Mutez)
	if !ok {
		return nil, nil
	}
	return m.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (m *Mutez) FromJSONSchema(data map[string]interface{}) error {
	setIntJSONSchema(&m.Default, data)
	return nil
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
func (b *Bool) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Bool)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	switch {
	case b.Value == secondItem.Value:
		return 0, nil
	case b.Value:
		return 1, nil
	default:
		return -1, nil
	}
}

// Distinguish -
func (b *Bool) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Bool)
	if !ok {
		return nil, nil
	}
	return b.Default.Distinguish(&second.Default)
}

// ToParameters -
func (b *Bool) ToParameters() ([]byte, error) {
	if v, ok := b.Value.(bool); ok && v {
		return []byte(`{"prim":"True"}`), nil
	}
	return []byte(`{"prim":"False"}`), nil
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
		return toBaseNodeString(val), nil
	case string:
		return toBaseNodeString(ts), nil
	}
	return nil, errors.Errorf("Invalid timestamp type")
}

// FromJSONSchema -
func (t *Timestamp) FromJSONSchema(data map[string]interface{}) error {
	for key := range data {
		if key == t.GetName() {
			t.ValueKind = valueKindInt
			switch val := data[key].(type) {
			case string:
				ts, err := time.Parse(time.RFC3339, val)
				if err != nil {
					return err
				}
				t.Value = base.NewBigInt(ts.UTC().Unix())
			case float64:
				t.Value = base.NewBigInt(int64(val))
			}
			break
		}
	}
	return nil
}

// ToParameters -
func (t *Timestamp) ToParameters() ([]byte, error) {
	switch ts := t.Value.(type) {
	case time.Time:
		return []byte(fmt.Sprintf(`{"int":"%d"}`, ts.UTC().Unix())), nil
	case *base.BigInt:
		return []byte(fmt.Sprintf(`{"int":"%d"}`, ts.Int64())), nil
	default:
		return nil, errors.Wrapf(consts.ErrInvalidType, "Timestamp.ToParameters: %T", t.Value)
	}
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
func (t *Timestamp) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Timestamp)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	ts := t.Value.(time.Time)
	ts2 := secondItem.Value.(time.Time)
	switch {
	case ts.Equal(ts2):
		return 0, nil
	case ts.Before(ts2):
		return -1, nil
	default:
		return 1, nil
	}
}

// Distinguish -
func (t *Timestamp) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Timestamp)
	if !ok {
		return nil, nil
	}
	return t.Default.Distinguish(&second.Default)
}

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
	setBytesJSONSchema(&b.Default, data)
	return nil
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
func (n *Never) Compare(second Comparable) (int, error) {
	if _, ok := second.(*Never); !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return 0, nil
}

// Distinguish -
func (n *Never) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Never)
	if !ok {
		return nil, nil
	}
	return n.Default.Distinguish(&second.Default)
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

// Distinguish -
func (o *Operation) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Operation)
	if !ok {
		return nil, nil
	}
	return o.Default.Distinguish(&second.Default)
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
	if a.ValueKind == valueKindBytes {
		return toBaseNodeBytes(val), nil
	}
	if optimized {
		value, err := forge.Contract(val)
		if err != nil {
			return nil, err
		}
		return toBaseNodeBytes(value), nil
	}
	return toBaseNodeString(val), nil
}

// ToMiguel -
func (a *Address) ToMiguel() (*MiguelNode, error) {
	name := a.GetTypeName()
	value := a.Value.(string)
	if a.ValueKind == valueKindBytes {
		v, err := forge.UnforgeAddress(value)
		if err != nil {
			return nil, err
		}
		value = v
	}
	return &MiguelNode{
		Prim:  a.Prim,
		Type:  strings.ToLower(a.Prim),
		Value: value,
		Name:  &name,
	}, nil
}

// ToJSONSchema -
func (a *Address) ToJSONSchema() (*JSONSchema, error) {
	return getAddressJSONSchema(a.Default), nil
}

// Compare -
func (a *Address) Compare(second Comparable) (int, error) {
	secondAddress, ok := second.(*Address)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	if a.Value == secondAddress.Value {
		return 0, nil
	}
	if a.ValueKind == secondAddress.ValueKind {
		return strings.Compare(a.Value.(string), secondAddress.Value.(string)), nil
	}
	return compareNotOptimizedTypes(a.Default, secondAddress.Default, forge.Contract)
}

// Distinguish -
func (a *Address) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Address)
	if !ok {
		return nil, nil
	}
	if err := a.optimizeStringValue(forge.UnforgeContract); err != nil {
		return nil, err
	}
	if err := second.optimizeStringValue(forge.UnforgeContract); err != nil {
		return nil, err
	}
	return a.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (a *Address) FromJSONSchema(data map[string]interface{}) error {
	setOptimizedJSONSchema(&a.Default, data, forge.UnforgeContract)
	return nil
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

// ToMiguel -
func (k *Key) ToMiguel() (*MiguelNode, error) {
	name := k.GetTypeName()
	value := k.Value.(string)
	if k.ValueKind == valueKindBytes {
		v, err := forge.UnforgePublicKey(value)
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
func (k *Key) ToBaseNode(optimized bool) (*base.Node, error) {
	val := k.Value.(string)
	if optimized {
		value, err := forge.PublicKey(val)
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
func (k *Key) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Key)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	return strings.Compare(k.Value.(string), secondItem.Value.(string)), nil
}

// Distinguish -
func (k *Key) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Key)
	if !ok {
		return nil, nil
	}
	return k.Default.Distinguish(&second.Default)
}

// FromJSONSchema -
func (k *Key) FromJSONSchema(data map[string]interface{}) error {
	setOptimizedJSONSchema(&k.Default, data, forge.UnforgePublicKey)
	return nil
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

func compareNotOptimizedTypes(x, y Default, optimizer func(string) (string, error)) (int, error) {
	if x.ValueKind != valueKindBytes {
		value, err := optimizer(x.Value.(string))
		if err != nil {
			return 0, err
		}
		x.ValueKind = valueKindBytes
		x.Value = value
	}
	if y.ValueKind != valueKindBytes {
		value, err := optimizer(y.Value.(string))
		if err != nil {
			return 0, err
		}
		y.ValueKind = valueKindBytes
		y.Value = value
	}

	return strings.Compare(x.Value.(string), y.Value.(string)), nil
}

func compareBigInt(x, y Default) int {
	xi := x.Value.(*base.BigInt)
	yi := y.Value.(*base.BigInt)
	return xi.Cmp(yi.Int)
}
