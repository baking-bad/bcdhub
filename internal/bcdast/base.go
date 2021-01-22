package bcdast

import (
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
)

// TODO: make unique annotations

// Default -
type Default struct {
	Prim      string
	TypeName  string
	FieldName string

	Value interface{}

	argsCount int
	annots    []string
	id        int
	depth     int
}

// NewDefault -
func NewDefault(prim string, argsCount, depth int) Default {
	return Default{
		Prim:      prim,
		argsCount: argsCount,
		depth:     depth,
	}
}

// MarshalJSON -
func (d Default) MarshalJSON() ([]byte, error) {
	return marshalJSON(d.Prim, d.annots)
}

// ParseType -
func (d *Default) ParseType(untyped Untyped, id *int) error {
	(*id)++
	d.id = *id
	d.annots = untyped.Annots
	d.FieldName = getAnnotation(untyped.Annots, prefixFieldName)
	d.TypeName = getAnnotation(untyped.Annots, prefixTypeName)

	prim := strings.ToLower(untyped.Prim)

	if prim != d.Prim {
		return errors.Wrap(ErrInvalidPrim, fmt.Sprintf("expected %s got %s", d.Prim, untyped.Prim))
	}

	if len(untyped.Args) != d.argsCount && d.argsCount >= 0 {
		return errors.Wrap(ErrInvalidArgsCount, fmt.Sprintf("expected %d got %d", d.argsCount, len(untyped.Args)))
	}

	switch {
	case untyped.IntValue != nil:
		d.Prim = consts.INT
	case untyped.StringValue != nil:
		d.Prim = consts.STRING
	case untyped.BytesValue != nil:
		d.Prim = consts.BYTES
	}
	return nil
}

// String -
func (d *Default) String() string {
	var typ string
	if d.FieldName != "" {
		typ = d.FieldName
	} else if d.TypeName != "" {
		typ = d.TypeName
	}
	if d.Value != nil {
		return fmt.Sprintf("[%d] %s=%v %s\n", d.id, d.Prim, d.Value, typ)
	}
	return fmt.Sprintf("[%d] %s %s\n", d.id, d.Prim, typ)
}

// ParseValue -
func (d *Default) ParseValue(untyped Untyped) error {
	switch {
	case untyped.IntValue != nil:
		d.Value = *untyped.IntValue
	case untyped.StringValue != nil:
		d.Value = *untyped.StringValue
	case untyped.BytesValue != nil:
		d.Value = *untyped.BytesValue
	}
	return nil
}

// ToMiguel -
func (d *Default) ToMiguel() (*MiguelNode, error) {
	return &MiguelNode{
		Prim:  d.Prim,
		Type:  strings.ToLower(d.Prim),
		Value: d.Value,
		Name:  d.GetName(),
	}, nil
}

// GetEntrypoints -
func (d *Default) GetEntrypoints() []string {
	switch {
	case d.TypeName != "":
		return []string{d.TypeName}
	case d.FieldName != "":
		return []string{d.FieldName}
	}
	return []string{""}
}

// GetName -
func (d *Default) GetName() string {
	switch {
	case d.TypeName != "":
		return d.TypeName
	case d.FieldName != "":
		return d.FieldName
	}
	return fmt.Sprintf("@%s_%d", d.Prim, d.id)
}

// GetValue -
func (d *Default) GetValue() interface{} {
	return d.Value
}

// GetPrim -
func (d *Default) GetPrim() string {
	return d.Prim
}

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
func (b *Bool) ParseValue(untyped Untyped) error {
	switch untyped.Prim {
	case consts.False:
		b.Value = false
	case consts.True:
		b.Value = true
	default:
		return ErrInvalidPrim
	}
	return nil
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
func (t *Timestamp) ParseValue(untyped Untyped) error {
	switch {
	case untyped.IntValue != nil:
		i := *untyped.IntValue
		if 253402300799 > i { // 31 December 9999 23:59:59 Golang time restriction
			t.Value = time.Unix(i, 0).UTC().String()
		} else {
			t.Value = fmt.Sprintf("%d", i)
		}
	case untyped.StringValue != nil:
		utc, err := time.Parse(time.RFC3339, *untyped.StringValue)
		if err != nil {
			return err
		}
		t.Value = utc.UTC().String()
	}
	return nil
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
