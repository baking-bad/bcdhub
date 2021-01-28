package ast

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/encoding"
	"github.com/baking-bad/bcdhub/internal/bcd/forge"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
)

// Default -
type Default struct {
	Prim      string
	TypeName  string
	FieldName string

	Value        interface{}
	UnforgeValue []*base.Node

	argsCount int
	annots    []string
	id        int
	depth     int
	valueType int
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
func (d *Default) ParseType(node *base.Node, id *int) error {
	(*id)++
	d.id = *id
	d.annots = node.Annots
	d.FieldName = getAnnotation(node.Annots, base.AnnotPrefixFieldName)
	d.TypeName = getAnnotation(node.Annots, base.AnnotPrefixrefixTypeName)

	prim := strings.ToLower(node.Prim)

	if prim != d.Prim {
		return errors.Wrap(base.ErrInvalidPrim, fmt.Sprintf("expected %s got %s", d.Prim, node.Prim))
	}

	if len(node.Args) != d.argsCount && d.argsCount >= 0 {
		return errors.Wrap(base.ErrInvalidArgsCount, fmt.Sprintf("expected %d got %d", d.argsCount, len(node.Args)))
	}

	switch {
	case node.IntValue != nil:
		d.Prim = consts.INT
	case node.StringValue != nil:
		d.Prim = consts.STRING
	case node.BytesValue != nil:
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
func (d *Default) ParseValue(node *base.Node) error {
	switch {
	case node.IntValue != nil:
		d.Value = node.IntValue
		d.valueType = valueTypeInt
	case node.StringValue != nil:
		d.Value = *node.StringValue
		d.valueType = valueTypeString
	case node.BytesValue != nil:
		d.Value = *node.BytesValue
		d.valueType = valueTypeBytes
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

// Forge -
func (d *Default) Forge(optimized bool) ([]byte, error) {
	switch d.valueType {
	case valueTypeString:
		val := d.Value.(string)
		return forgeString(val)
	case valueTypeBytes:
		val := d.Value.(string)
		return forgeBytes(val)
	case valueTypeInt:
		val := d.Value.(*base.BigInt)
		return forgeInt(val)
	default:
		forger := forge.NewMichelson()
		node := &base.Node{
			Prim:   d.Prim,
			Annots: d.annots,
			Args:   make([]*base.Node, 0),
		}
		forger.Nodes = []*base.Node{node}
		return forger.Forge()
	}
}

// Unforge -
func (d *Default) Unforge(data []byte) (int, error) {
	if d.valueType != valueTypeBytes {
		return 0, nil
	}
	unforger := forge.NewMichelson()
	n, err := unforger.UnforgeString(d.Value.(string))
	if err != nil {
		return n, err
	}
	d.UnforgeValue = unforger.Nodes
	return n, err
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
func (b *Bool) ParseValue(node *base.Node) error {
	switch node.Prim {
	case consts.False:
		b.Value = false
	case consts.True:
		b.Value = true
	default:
		return base.ErrInvalidPrim
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
func (t *Timestamp) ParseValue(node *base.Node) error {
	switch {
	case node.IntValue != nil:
		i := node.IntValue.Int64()
		if 253402300799 > i { // 31 December 9999 23:59:59 Golang time restriction
			t.Value = time.Unix(i, 0).UTC().String()
		} else {
			t.Value = fmt.Sprintf("%d", i)
		}
	case node.StringValue != nil:
		utc, err := time.Parse(time.RFC3339, *node.StringValue)
		if err != nil {
			return err
		}
		t.Value = utc.UTC().String()
	}
	return nil
}

// Forge -
func (t *Timestamp) Forge(optimized bool) ([]byte, error) {
	ts := t.Value.(time.Time)

	if optimized {
		return forgeInt(base.NewBigInt(ts.UTC().Unix()))
	}
	val := ts.UTC().Format(time.RFC3339)
	return forgeString(val)
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

// Forge -
func (c *ChainID) Forge(optimized bool) ([]byte, error) {
	val := c.Value.(string)
	if optimized {
		value, err := encoding.DecodeBase58ToString(val)
		if err != nil {
			return nil, err
		}
		return forgeBytes(value)
	}
	return forgeString(val)
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

// Forge -
func (a *Address) Forge(optimized bool) ([]byte, error) {
	val := a.Value.(string)
	if optimized {
		value, err := forgeContract(val)
		if err != nil {
			return nil, err
		}
		return forgeBytes(value)
	}
	return forgeString(val)
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

// Forge -
func (k *Key) Forge(optimized bool) ([]byte, error) {
	val := k.Value.(string)
	if optimized {
		value, err := forgePublicKey(val)
		if err != nil {
			return nil, err
		}
		s := hex.EncodeToString(value)
		return forgeBytes(s)
	}
	return forgeString(val)
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

// Forge -
func (k *KeyHash) Forge(optimized bool) ([]byte, error) {
	val := k.Value.(string)
	if optimized {
		value, err := forgeAddress(val, true)
		if err != nil {
			return nil, err
		}
		s := hex.EncodeToString(value)
		return forgeBytes(s)
	}
	return forgeString(val)
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

// Forge -
func (s *Signature) Forge(optimized bool) ([]byte, error) {
	val := s.Value.(string)
	if optimized {
		value, err := encoding.DecodeBase58ToString(val)
		if err != nil {
			return nil, err
		}
		return forgeBytes(value)
	}
	return forgeString(val)
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
