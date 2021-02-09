package ast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

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
	d.FieldName = getAnnotation(node.Annots, consts.AnnotPrefixFieldName)
	d.TypeName = getAnnotation(node.Annots, consts.AnnotPrefixrefixTypeName)

	prim := strings.ToLower(node.Prim)

	if prim != d.Prim {
		return errors.Wrap(consts.ErrInvalidPrim, fmt.Sprintf("expected %s got %s", d.Prim, node.Prim))
	}

	if len(node.Args) != d.argsCount && d.argsCount >= 0 {
		return errors.Wrap(consts.ErrInvalidArgsCount, fmt.Sprintf("expected %d got %d", d.argsCount, len(node.Args)))
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
	name := d.GetName()
	return &MiguelNode{
		Prim:  d.Prim,
		Type:  strings.ToLower(d.Prim),
		Value: d.Value,
		Name:  &name,
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

// IsNamed -
func (d *Default) IsNamed() bool {
	return d.FieldName != "" || d.TypeName != ""
}

// GetName -
func (d *Default) GetName() string {
	switch {
	case d.FieldName != "":
		return d.FieldName
	case d.TypeName != "":
		return d.TypeName
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

// IsPrim -
func (d *Default) IsPrim(prim string) bool {
	return d.Prim == prim
}

// ToBaseNode -
func (d *Default) ToBaseNode(optimized bool) (*base.Node, error) {
	node := new(base.Node)
	switch d.valueType {
	case valueTypeString:
		val := d.Value.(string)
		node.StringValue = &val
	case valueTypeBytes:
		val := d.Value.(string)
		node.BytesValue = &val
	case valueTypeInt:
		val := d.Value.(*base.BigInt)
		node.IntValue = val
	default:
		node.Prim = d.Prim
		node.Annots = d.annots
		node.Args = make([]*base.Node, 0)
	}
	return node, nil
}

// ToJSONSchema -
func (d *Default) ToJSONSchema() (*JSONSchema, error) {
	return wrapObject(&JSONSchema{
		Prim: d.Prim,
	}), nil
}

// FromJSONSchema -
func (d *Default) FromJSONSchema(data map[string]interface{}) error {
	for key := range data {
		if key == d.GetName() {
			d.Value = data[key]
			break
		}
	}
	return nil
}

// EnrichBigMap -
func (d *Default) EnrichBigMap(bmd []*base.BigMapDiff) error {
	return nil
}

// ToParameters -
func (d *Default) ToParameters() ([]byte, error) {
	switch d.valueType {
	case valueTypeString:
		return []byte(fmt.Sprintf(`{"string": "%s"}`, d.Value)), nil
	case valueTypeBytes:
		return []byte(fmt.Sprintf(`{"bytes": "%s"}`, d.Value)), nil
	case valueTypeInt:
		i := d.Value.(*base.BigInt)
		return []byte(fmt.Sprintf(`{"int": "%d"}`, i.Int64())), nil
	}
	return nil, nil
}

// Docs -
func (d *Default) Docs(inferredName string) ([]Typedef, string, error) {
	return []Typedef{
		{
			Name: d.GetName(),
			Type: d.Prim,
		},
	}, d.Prim, nil
}

// FindByName -
func (d *Default) FindByName(name string) Node {
	if d.GetName() == name {
		return d
	}
	return nil
}

// Distinguish -
func (d *Default) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Default)
	if !ok {
		return nil, nil
	}
	if d.Prim != second.Prim {
		return nil, errors.Wrapf(consts.ErrInvalidPrim, "%s != %s", d.Prim, second.Prim)
	}
	name := d.GetName()
	node := new(MiguelNode)
	node.Prim = d.Prim
	node.Type = d.Prim
	node.Name = &name
	node.Value = second.Value
	switch {
	case d.Value != nil && second.Value == nil:
		node.DiffType = MiguelKindDelete
		node.From = d.Value
	case d.Value == nil && second.Value != nil:
		node.DiffType = MiguelKindCreate
		node.From = d.Value
	case d.Value != nil && second.Value != nil && d.Value != second.Value:
		node.DiffType = MiguelKindUpdate
	}
	return node, nil
}

// Compare -
func (d *Default) Compare(second Comparable) (bool, error) {
	return false, consts.ErrTypeIsNotComparable
}

// Equal -
func (d *Default) Equal(second Node) bool {
	return d.Prim == second.GetPrim() && d.GetName() == second.GetName() && d.GetValue() == second.GetValue()
}
