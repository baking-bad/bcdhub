package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/types"
	"github.com/pkg/errors"
)

// Option -
type Option struct {
	Default
	Type Node
}

// NewOption -
func NewOption(depth int) *Option {
	return &Option{
		Default: NewDefault(consts.OPTION, 1, depth),
	}
}

// String -
func (opt *Option) String() string {
	var s strings.Builder
	s.WriteString(opt.Default.String())
	s.WriteString(strings.Repeat(consts.DefaultIndent, opt.Depth))
	s.WriteString(opt.Type.String())
	return s.String()
}

// MarshalJSON -
func (opt *Option) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.OPTION, opt.Annots, opt.Type)
}

// ParseType -
func (opt *Option) ParseType(node *base.Node, id *int) error {
	if err := opt.Default.ParseType(node, id); err != nil {
		return err
	}

	child, err := typingNode(node.Args[0], opt.Depth, id)
	if err != nil {
		return err
	}
	opt.Type = child

	return nil
}

// ParseValue -
func (opt *Option) ParseValue(node *base.Node) error {
	if len(node.Args) > opt.ArgsCount {
		return errors.Wrap(consts.ErrTreesAreDifferent, "Option.ParseValue")
	}

	switch node.Prim {
	case consts.None:
		opt.Value = consts.None
		return nil
	case consts.Some:
		opt.Value = consts.Some
		err := opt.Type.ParseValue(node.Args[0])
		return err
	default:
		return consts.ErrInvalidPrim
	}

}

// ToMiguel -
func (opt *Option) ToMiguel() (*MiguelNode, error) {
	var ast Node

	if opt.Value == consts.None {
		ast = &opt.Default
	} else {
		ast = opt.Type
	}

	node, err := ast.ToMiguel()
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(*node.Name, "@") {
		name := opt.GetTypeName()
		if !strings.HasPrefix(name, "@") {
			node.Name = &name
		}
	}
	return node, nil
}

// ToBaseNode -
func (opt *Option) ToBaseNode(optimized bool) (*base.Node, error) {
	node := new(base.Node)

	if opt.Value == consts.None {
		node.Prim = consts.None
	} else {
		node.Prim = consts.Some
		arg, err := opt.Type.ToBaseNode(optimized)
		if err != nil {
			return nil, err
		}
		node.Args = []*base.Node{arg}
	}

	return node, nil
}

var noneSchema = &JSONSchema{
	Type:  JSONSchemaTypeString,
	Const: consts.NONE,
}

// ToJSONSchema -
func (opt *Option) ToJSONSchema() (*JSONSchema, error) {
	someSchema := &JSONSchema{
		Title:      consts.Some,
		Properties: make(map[string]*JSONSchema),
	}

	someProperties := &JSONSchema{
		Type:  JSONSchemaTypeString,
		Const: consts.SOME,
	}

	child, err := opt.Type.ToJSONSchema()
	if err != nil {
		return nil, err
	}

	if len(child.Properties) > 0 {
		for key, value := range child.Properties {
			someSchema.Properties[key] = value
		}
	}

	someSchema.Properties["schemaKey"] = someProperties

	return &JSONSchema{
		Type:  JSONSchemaTypeObject,
		Prim:  opt.Prim,
		Title: opt.GetName(),
		OneOf: []*JSONSchema{
			{
				Title: consts.None,
				Properties: map[string]*JSONSchema{
					"schemaKey": noneSchema,
				},
			},
			someSchema,
		},
		Default: &JSONSchema{
			SchemaKey: (*SchemaKey)(noneSchema),
		},
	}, nil
}

// FromJSONSchema -
func (opt *Option) FromJSONSchema(data map[string]interface{}) error {
	var optionMap map[string]interface{}
	for key := range data {
		if key == opt.GetName() {
			val := data[key]
			arrVal, ok := val.(map[string]interface{})
			if !ok {
				return errors.Wrapf(consts.ErrInvalidType, "Option.FromJSONSchema %T", val)
			}
			optionMap = arrVal
			break
		}
	}
	schemaKey, ok := optionMap["schemaKey"]
	if !ok {
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "Option.FromJSONSchema")
	}
	delete(optionMap, "schemaKey")

	switch schemaKey {
	case consts.NONE:
		opt.Value = consts.None
	case consts.SOME:
		return opt.Type.FromJSONSchema(optionMap)
	default:
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "Option.FromJSONSchema")
	}
	return nil
}

// EnrichBigMap -
func (opt *Option) EnrichBigMap(bmd []*types.BigMapDiff) error {
	if opt.Type != nil {
		return opt.Type.EnrichBigMap(bmd)
	}
	return nil
}

// ToParameters -
func (opt *Option) ToParameters() ([]byte, error) {
	if opt.Value == consts.None {
		return []byte(`{"prim":"None"}`), nil
	}

	var builder bytes.Buffer
	if _, err := builder.WriteString(`{"prim":"Some","args":[`); err != nil {
		return nil, err
	}

	b, err := opt.Type.ToParameters()
	if err != nil {
		return nil, err
	}
	if _, err := builder.Write(b); err != nil {
		return nil, err
	}

	if _, err := builder.WriteString(`]}`); err != nil {
		return nil, err
	}
	return builder.Bytes(), nil
}

// FindByName -
func (opt *Option) FindByName(name string) Node {
	if opt.GetName() == name {
		return opt
	}
	return opt.Type.FindByName(name)
}

// Docs -
func (opt *Option) Docs(inferredName string) ([]Typedef, string, error) {
	name := getNameDocString(opt, inferredName)
	docs, varName, err := opt.Type.Docs(name)
	if err != nil {
		return nil, "", err
	}

	optName := fmt.Sprintf("option(%s)", varName)
	if isSimpleDocType(docs[0].Type) {
		return nil, optName, nil
	}
	return docs, optName, nil
}

// Compare -
func (opt *Option) Compare(second Comparable) (int, error) {
	secondItem, ok := second.(*Option)
	if !ok {
		return 0, consts.ErrTypeIsNotComparable
	}
	switch {
	case opt.Value == consts.None && secondItem.Value == consts.None:
		return 0, nil
	case opt.Value != secondItem.Value:
		if opt.Value == consts.Some {
			return 1, nil
		}
		return -1, nil
	default:
		return opt.Type.Compare(secondItem.Type)
	}
}

// Distinguish -
func (opt *Option) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Option)
	if !ok {
		return nil, nil
	}

	switch {
	case opt.Value == consts.None && second.Value == consts.None:
		name := opt.getTypeName()
		return &MiguelNode{
			Name:  &name,
			Value: opt.Value,
		}, nil
	case opt.Value == consts.None && second.Value == consts.Some:
		name := opt.getTypeName()
		node, err := second.Type.ToMiguel()
		if err != nil {
			return nil, err
		}
		node.Name = &name
		node.setDiffType(MiguelKindDelete)
		return node, nil
	case opt.Value == consts.Some && second.Value == consts.None:
		name := opt.getTypeName()
		node, err := opt.Type.ToMiguel()
		if err != nil {
			return nil, err
		}
		node.Name = &name
		node.setDiffType(MiguelKindCreate)
		return node, nil
	case opt.Value == consts.Some && second.Value == consts.Some:
		child, err := opt.Type.Distinguish(second.Type)
		if err != nil {
			return nil, err
		}
		return child, err
	}

	return nil, nil
}

// EqualType -
func (opt *Option) EqualType(node Node) bool {
	if !opt.Default.EqualType(node) {
		return false
	}
	second, ok := node.(*Option)
	if !ok {
		return false
	}

	return opt.Type.EqualType(second.Type)
}

func (opt *Option) getTypeName() string {
	name := opt.Type.GetTypeName()
	if !strings.HasPrefix(name, "@") {
		return name
	}
	optName := opt.GetTypeName()
	if !strings.HasPrefix(optName, "@") {
		return optName
	}
	return name
}

// FindPointers -
func (opt *Option) FindPointers() map[int64]*BigMap {
	if opt.Value == consts.SOME {
		return opt.Type.FindPointers()
	}
	return nil
}

// Range -
func (opt *Option) Range(handler func(node Node) error) error {
	if opt.Value == consts.SOME {
		return opt.Type.Range(handler)
	}
	return nil
}

// GetJSONModel -
func (opt *Option) GetJSONModel(model JSONModel) {
	if model == nil {
		return
	}
	item := JSONModel{
		"schemaKey": opt.Value,
	}
	if opt.Value == consts.SOME {
		opt.Type.GetJSONModel(item)
	}

	model[opt.GetName()] = item
}
