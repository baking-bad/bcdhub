package ast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
)

// Default -
type Default struct {
	Prim      string
	TypeName  string
	FieldName string

	Value     interface{}
	ValueKind int
	ID        int

	ArgsCount int
	Depth     int
	Annots    []string
}

// NewDefault -
func NewDefault(prim string, argsCount, depth int) Default {
	return Default{
		Prim:      prim,
		ArgsCount: argsCount,
		Depth:     depth,
	}
}

// MarshalJSON -
func (d Default) MarshalJSON() ([]byte, error) {
	return marshalJSON(d.Prim, d.Annots)
}

// ParseType -
func (d *Default) ParseType(node *base.Node, id *int) error {
	(*id)++
	d.ID = *id
	d.Annots = node.Annots
	d.FieldName = getAnnotation(node.Annots, consts.AnnotPrefixFieldName)
	d.TypeName = getAnnotation(node.Annots, consts.AnnotPrefixrefixTypeName)

	prim := strings.ToLower(node.Prim)

	if prim != d.Prim {
		return errors.Wrap(consts.ErrInvalidPrim, fmt.Sprintf("expected %s got %s", d.Prim, node.Prim))
	}

	if len(node.Args) != d.ArgsCount && d.ArgsCount >= 0 {
		return errors.Wrap(consts.ErrInvalidArgsCount, fmt.Sprintf("expected %d got %d", d.ArgsCount, len(node.Args)))
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
		return fmt.Sprintf("[%d] %s=%v %s\n", d.ID, d.Prim, d.Value, typ)
	}
	return fmt.Sprintf("[%d] %s %s\n", d.ID, d.Prim, typ)
}

// ParseValue -
func (d *Default) ParseValue(node *base.Node) error {
	switch {
	case node.IntValue != nil:
		d.Value = node.IntValue
		d.ValueKind = valueKindInt
	case node.StringValue != nil:
		d.Value = *node.StringValue
		d.ValueKind = valueKindString
	case node.BytesValue != nil:
		d.Value = *node.BytesValue
		d.ValueKind = valueKindBytes
	}
	return nil
}

// GetTypeName -
func (d *Default) GetTypeName() string {
	switch {
	case d.TypeName != "":
		return d.TypeName
	case d.FieldName != "":
		return d.FieldName
	default:
		return fmt.Sprintf("@%s_%d", d.Prim, d.ID)
	}
}

// ToMiguel -
func (d *Default) ToMiguel() (*MiguelNode, error) {
	name := d.GetTypeName()
	node := &MiguelNode{
		Prim: d.Prim,
		Type: strings.ToLower(d.Prim),
		Name: &name,
	}
	node.Value = d.miguelValue()

	return node, nil
}

// GetEntrypoints -
func (d *Default) GetEntrypoints() []string {
	if d.FieldName != "" {
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
	return fmt.Sprintf("@%s_%d", d.Prim, d.ID)
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
	switch d.ValueKind {
	case valueKindString:
		val := d.Value.(string)
		node.StringValue = &val
	case valueKindBytes:
		val := d.Value.(string)
		node.BytesValue = &val
	case valueKindInt:
		val := d.Value.(*types.BigInt)
		node.IntValue = val
	default:
		node.Prim = d.Prim
		node.Annots = d.Annots
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
			d.ValueKind = valueKindString
			break
		}
	}
	return nil
}

// EnrichBigMap -
func (d *Default) EnrichBigMap(bmd []*types.BigMapDiff) error {
	return nil
}

// ToParameters -
func (d *Default) ToParameters() ([]byte, error) {
	switch d.ValueKind {
	case valueKindString:
		return []byte(fmt.Sprintf(`{"string":"%s"}`, d.Value)), nil
	case valueKindBytes:
		return []byte(fmt.Sprintf(`{"bytes":"%s"}`, d.Value)), nil
	case valueKindInt:
		i := d.Value.(*types.BigInt)
		return []byte(fmt.Sprintf(`{"int":"%d"}`, i.Int64())), nil
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
	node.Value = d.miguelValue()

	switch {
	case d.Value != nil && second.Value == nil:
		node.DiffType = MiguelKindDelete
	case d.Value == nil && second.Value != nil:
		node.DiffType = MiguelKindCreate
	case d.Value != nil && second.Value != nil:
		switch v := d.Value.(type) {
		case *types.BigInt:
			sv := second.Value.(*types.BigInt)
			if sv.Cmp(v.Int) == 0 {
				return node, nil
			}
			node.From = sv.String()
		default:
			if d.Value == second.Value {
				return node, nil
			}
			node.From = second.miguelValue()
		}
		node.DiffType = MiguelKindUpdate
	}
	return node, nil
}

// Compare -
func (d *Default) Compare(second Comparable) (int, error) {
	return 0, consts.ErrTypeIsNotComparable
}

// EqualType -
func (d *Default) EqualType(node Node) bool {
	return d.Prim == node.GetPrim()
}

// Equal -
func (d *Default) Equal(node Node) bool {
	return d.EqualType(node) && d.GetName() == node.GetName() && d.equalValue(node.GetValue())
}

func (d *Default) equalValue(value interface{}) bool {
	switch {
	case d.Value == nil && value == nil:
		return true
	case d.Value == nil && value != nil:
		return false
	case d.Value != nil && value == nil:
		return false
	default:
		switch val := d.Value.(type) {
		case *types.BigInt:
			if sv, ok := value.(*types.BigInt); ok {
				return val.Cmp(sv.Int) == 0
			}
			return false
		default:
			return val == value
		}
	}
}

func (d *Default) miguelValue() interface{} {
	if d.Value == nil {
		return nil
	}
	switch v := d.Value.(type) {
	case *types.BigInt:
		return v.String()
	default:
		return d.Value
	}
}

func (d *Default) optimizeStringValue(optimizer func(string) (string, error)) error {
	if d.ValueKind != valueKindBytes {
		return nil
	}
	sv := d.Value.(string)
	newValue, err := optimizer(sv)
	if err != nil {
		return err
	}
	d.Value = newValue
	return nil
}

// FindPointers -
func (d *Default) FindPointers() map[int64]*BigMap {
	return nil
}

// Range -
func (d *Default) Range(handler func(node Node) error) error {
	return handler(d)
}

// GetJSONModel -
func (d *Default) GetJSONModel(model JSONModel) {
	if model == nil {
		return
	}
	model[d.GetName()] = d.Value
}
