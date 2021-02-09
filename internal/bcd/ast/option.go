package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

// Option -
type Option struct {
	Default
	Child Node
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
	s.WriteString(strings.Repeat(consts.DefaultIndent, opt.depth))
	s.WriteString(opt.Child.String())
	return s.String()
}

// MarshalJSON -
func (opt *Option) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.OPTION, opt.annots, opt.Child)
}

// ParseType -
func (opt *Option) ParseType(node *base.Node, id *int) error {
	if err := opt.Default.ParseType(node, id); err != nil {
		return err
	}

	child, err := typingNode(node.Args[0], opt.depth, id)
	if err != nil {
		return err
	}
	opt.Child = child

	return nil
}

// ParseValue -
func (opt *Option) ParseValue(node *base.Node) error {
	if len(node.Args) > opt.argsCount {
		return errors.Wrap(consts.ErrTreesAreDifferent, "Option.ParseValue")
	}

	switch node.Prim {
	case consts.None:
		opt.Value = nil
		return nil
	case consts.Some:
		opt.Value = consts.Some
		err := opt.Child.ParseValue(node.Args[0])
		return err
	default:
		return consts.ErrInvalidPrim
	}

}

// ToMiguel -
func (opt *Option) ToMiguel() (*MiguelNode, error) {
	var ast Node

	if opt.Value == nil {
		ast = &opt.Default
	} else {
		ast = opt.Child
	}

	return ast.ToMiguel()
}

// ToBaseNode -
func (opt *Option) ToBaseNode(optimized bool) (*base.Node, error) {
	node := new(base.Node)

	if opt.Value == nil {
		node.Prim = consts.None
	} else {
		node.Prim = consts.Some
		arg, err := opt.Child.ToBaseNode(optimized)
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

	child, err := opt.Child.ToJSONSchema()
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
		opt.Value = nil
	case consts.SOME:
		val, err := createByType(opt.Child)
		if err != nil {
			return err
		}
		var ok bool
		for key := range optionMap {
			if err := val.FromJSONSchema(optionMap[key].(map[string]interface{})); err == nil {
				ok = true
				break
			}
		}
		if ok {
			opt.Value = val
		} else {
			return errors.Wrap(consts.ErrJSONDataIsAbsent, "Option.FromJSONSchema")
		}
	default:
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "Option.FromJSONSchema")
	}
	return nil
}

// EnrichBigMap -
func (opt *Option) EnrichBigMap(bmd []*base.BigMapDiff) error {
	if opt.Child != nil {
		return opt.Child.EnrichBigMap(bmd)
	}
	return nil
}

// ToParameters -
func (opt *Option) ToParameters() ([]byte, error) {
	if opt.Value == nil {
		return []byte(`{"prim":"None"}`), nil
	}

	var builder bytes.Buffer
	if _, err := builder.WriteString(`{"prim":"Some","args":[`); err != nil {
		return nil, err
	}

	b, err := opt.Child.ToParameters()
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
	return opt.Child.FindByName(name)
}

// Docs -
func (opt *Option) Docs(inferredName string) ([]Typedef, string, error) {
	name := getNameDocString(opt, inferredName)
	docs, varName, err := opt.Child.Docs(name)
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
func (opt *Option) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*Option)
	if !ok {
		return false, nil
	}
	switch {
	case opt.Value == nil && secondItem.Value == nil:
		return true, nil
	case opt.Value != secondItem.Value:
		return false, nil
	default:
		return opt.Child.Compare(secondItem.Child)
	}
}

// Distinguish -
func (opt *Option) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Option)
	if !ok {
		return nil, nil
	}

	switch {
	case opt.Value == nil && second.Value == nil:
		return opt.ToMiguel()
	case opt.Value == nil && second.Value != nil:
		node, err := second.ToMiguel()
		if err != nil {
			return nil, err
		}
		node.DiffType = MiguelKindCreate
		return node, nil
	case opt.Value != nil && second.Value == nil:
		node, err := opt.ToMiguel()
		if err != nil {
			return nil, err
		}
		node.DiffType = MiguelKindDelete
		return node, nil
	case opt.Value != nil && second.Value != nil:
		child, err := opt.Distinguish(second.Child)
		if err != nil {
			return nil, err
		}
		child.DiffType = MiguelKindUpdate
		return child, err
	}

	return nil, nil
}
