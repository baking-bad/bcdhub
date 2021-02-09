package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/pkg/errors"
)

// Or -
type Or struct {
	Default
	Args []Node

	Left  Node
	Right Node
}

// NewOr -
func NewOr(depth int) *Or {
	return &Or{
		Default: NewDefault(consts.OR, -1, depth),
	}
}

// String -
func (or *Or) String() string {
	var s strings.Builder
	switch {
	case or.Left != nil:
		s.WriteString(consts.Left)
		s.WriteByte(' ')
		s.WriteString(or.Default.String())
		s.WriteString(strings.Repeat(consts.DefaultIndent, or.depth))
		s.WriteString(or.Left.String())
		s.WriteByte(' ')
	case or.Right != nil:
		s.WriteString(consts.Right)
		s.WriteByte(' ')
		s.WriteString(or.Default.String())
		s.WriteString(strings.Repeat(consts.DefaultIndent, or.depth))
		s.WriteString(or.Right.String())
		s.WriteByte(' ')
	default:
		s.WriteString(or.Default.String())
		for i := range or.Args {
			s.WriteString(strings.Repeat(consts.DefaultIndent, or.depth))
			s.WriteString(or.Args[i].String())
		}

	}
	return s.String()
}

// MarshalJSON -
func (or *Or) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.OR, or.annots, or.Args...)
}

// ParseType -
func (or *Or) ParseType(node *base.Node, id *int) error {
	if err := or.Default.ParseType(node, id); err != nil {
		return err
	}

	if len(node.Args) > 2 || len(node.Args) == 0 {
		return errors.Wrap(consts.ErrInvalidArgsCount, "Or.ParseValue")
	}

	or.Args = make([]Node, 0, len(node.Args))
	for _, arg := range node.Args {
		child, err := typingNode(arg, or.depth, id)
		if err != nil {
			return err
		}
		or.Args = append(or.Args, child)
	}

	return nil
}

// ParseValue -
func (or *Or) ParseValue(node *base.Node) error {
	if len(node.Args) > 2 || len(node.Args) == 0 {
		return errors.Wrap(consts.ErrInvalidArgsCount, "Or.ParseValue")
	}

	switch node.Prim {
	case consts.Left:
		if err := or.Args[0].ParseValue(node.Args[0]); err != nil {
			return err
		}
		or.Left = or.Args[0]
	case consts.Right:
		if err := or.Args[1].ParseValue(node.Args[0]); err != nil {
			return err
		}
		or.Right = or.Args[1]
	default:
		return errors.Wrap(consts.ErrInvalidPrim, "Or.ParseValue")
	}
	return nil
}

// ToMiguel -
func (or *Or) ToMiguel() (*MiguelNode, error) {
	node, err := or.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	node.Children = make([]*MiguelNode, 0)
	for i := range or.Args {
		child, err := or.Args[i].ToMiguel()
		if err != nil {
			return nil, err
		}

		if or.Prim == or.Args[i].GetPrim() {
			node.Children = append(node.Children, child.Children...)
		} else {
			node.Children = append(node.Children, child)
		}
	}

	node.Type = consts.TypeNamedEnum
	for i := range node.Children {
		if node.Children[i].Prim != consts.UNIT {
			node.Type = consts.TypeNamedUnion
			break
		}
	}

	return node, nil
}

// GetEntrypoints -
func (or *Or) GetEntrypoints() []string {
	e := make([]string, 0)
	for i := range or.Args {
		e = append(e, or.Args[i].GetEntrypoints()...)
	}
	return e
}

// ToBaseNode -
func (or *Or) ToBaseNode(optimized bool) (*base.Node, error) {
	node := new(base.Node)
	switch {
	case or.Left != nil:
		node.Prim = consts.Left
		arg, err := or.Left.ToBaseNode(optimized)
		if err != nil {
			return nil, err
		}
		node.Args = []*base.Node{arg}
	case or.Right != nil:
		node.Prim = consts.Right
		arg, err := or.Right.ToBaseNode(optimized)
		if err != nil {
			return nil, err
		}
		node.Args = []*base.Node{arg}
	default:
		return nil, errors.New("OR is not Left or Right")
	}
	return node, nil
}

// ToJSONSchema -
func (or *Or) ToJSONSchema() (*JSONSchema, error) {
	oneOf := make([]*JSONSchema, 0)
	for i, arg := range or.Args {
		child, err := arg.ToJSONSchema()
		if err != nil {
			return nil, err
		}

		if child.Prim == consts.OR {
			oneOf = append(oneOf, child.OneOf...)
		} else {
			key := consts.LEFT
			if i == 1 {
				key = consts.RIGHT
			}
			item := &JSONSchema{
				Title: arg.GetName(),
				Properties: map[string]*JSONSchema{
					"schemaKey": {
						Type:  JSONSchemaTypeString,
						Const: key,
					},
				},
			}

			if child.Prim != consts.UNIT {
				for key, value := range child.Properties {
					item.Properties[key] = value
				}
			}

			oneOf = append(oneOf, item)
		}
	}

	return &JSONSchema{
		Type:  JSONSchemaTypeObject,
		Title: or.GetName(),
		Prim:  or.Prim,
		OneOf: oneOf,
	}, nil
}

// FromJSONSchema -
func (or *Or) FromJSONSchema(data map[string]interface{}) error {
	var orMap map[string]interface{}
	for key := range data {
		if key == or.GetName() {
			val := data[key]
			arrVal, ok := val.(map[string]interface{})
			if !ok {
				return errors.Wrapf(consts.ErrInvalidType, "Or.FromJSONSchema %T", val)
			}
			orMap = arrVal
			break
		}
	}
	schemaKey, ok := orMap["schemaKey"]
	if !ok {
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "Or.FromJSONSchema")
	}
	delete(orMap, "schemaKey")

	switch schemaKey {
	case consts.LEFT:
		val, err := createByType(or.Args[0])
		if err != nil {
			return err
		}
		if err := val.FromJSONSchema(orMap); err != nil {
			return err
		}
		or.Left = val
	case consts.RIGHT:
		val, err := createByType(or.Args[1])
		if err != nil {
			return err
		}
		if err := val.FromJSONSchema(orMap); err != nil {
			return err
		}
		or.Right = val
	default:
		return errors.Wrap(consts.ErrJSONDataIsAbsent, "Or.FromJSONSchema")
	}

	return nil
}

// EnrichBigMap -
func (or *Or) EnrichBigMap(bmd []*base.BigMapDiff) error {
	if or.Left != nil {
		if err := or.Left.EnrichBigMap(bmd); err != nil {
			return err
		}
	}
	if or.Right != nil {
		if err := or.Right.EnrichBigMap(bmd); err != nil {
			return err
		}
	}

	return nil
}

// ToParameters -
func (or *Or) ToParameters() ([]byte, error) {
	var builder bytes.Buffer

	var prim string
	var node Node
	switch {
	case or.Left != nil:
		prim = consts.Left
		node = or.Left
	case or.Right != nil:
		prim = consts.Right
		node = or.Right
	}

	if prim != "" {
		if _, err := builder.WriteString(fmt.Sprintf(`{"prim":"%s","args":[`, prim)); err != nil {
			return nil, err
		}
		b, err := node.ToParameters()
		if err != nil {
			return nil, err
		}
		if _, err := builder.Write(b); err != nil {
			return nil, err
		}
		if _, err := builder.WriteString(`]}`); err != nil {
			return nil, err
		}
	}
	return builder.Bytes(), nil
}

// FindByName -
func (or *Or) FindByName(name string) Node {
	if or.GetName() == name {
		return or
	}
	for i := range or.Args {
		node := or.Args[i].FindByName(name)
		if node != nil {
			return node
		}
	}
	return nil
}

// Docs -
func (or *Or) Docs(inferredName string) ([]Typedef, string, error) {
	name := getNameDocString(or, inferredName)

	typedef := Typedef{
		Name: name,
		Type: or.Prim,
		Args: make([]TypedefArg, 0),
	}
	result := make([]Typedef, 0)
	for i := range or.Args {
		if isSimpleDocType(or.Args[i].GetPrim()) {
			typedef.Args = append(typedef.Args, TypedefArg{
				Key:   or.Args[i].GetName(),
				Value: or.Args[i].GetPrim(),
			})
			continue
		}

		args, varName, err := or.Args[i].Docs(name)
		if err != nil {
			return nil, "", err
		}
		if or.Args[i].IsPrim(or.Prim) {
			typedef.Args = append(typedef.Args, args[0].Args...)
			for j := range args {
				if args[j].Type != or.Prim {
					result = append(result, args[j])
				}
			}
		} else {
			typedef.Args = append(typedef.Args, TypedefArg{
				Key:   or.Args[i].GetName(),
				Value: varName,
			})
			for j := range args {
				if !isFlatDocType(args[j]) {
					result = append(result, args[j])
				}
			}
		}
	}
	result = append([]Typedef{typedef}, result...)
	return result, makeVarDocString(name), nil
}

// Compare -
func (or *Or) Compare(second Comparable) (bool, error) {
	secondItem, ok := second.(*Or)
	if !ok {
		return false, nil
	}
	var err error
	switch {
	case or.Left != nil && secondItem.Left != nil:
		ok, err = or.Left.Compare(secondItem.Left)
	case or.Right != nil && secondItem.Right != nil:
		ok, err = or.Right.Compare(secondItem.Right)
	default:
		return false, nil
	}
	if err != nil {
		if errors.Is(err, consts.ErrTypeIsNotComparable) {
			return false, nil
		}
		return false, err
	}
	return ok, nil
}

// Distinguish -
func (or *Or) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Or)
	if !ok {
		return nil, nil
	}

	node, err := or.Default.ToMiguel()
	if err != nil {
		return nil, err
	}
	node.Children = make([]*MiguelNode, 0)

	switch {
	case second.Left != nil && or.Left != nil:
		child, err := or.Left.Distinguish(second.Left)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	case second.Right != nil && or.Right != nil:
		child, err := or.Right.Distinguish(second.Right)
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	case second.Right == nil && or.Right == nil && second.Left == nil && or.Left == nil:
	default:
		if second.Right != nil {
			child, err := second.Right.ToMiguel()
			if err != nil {
				return nil, err
			}
			node.DiffType = MiguelKindUpdate
			node.Children = append(node.Children, child)
		}
		if second.Left != nil {
			child, err := second.Left.ToMiguel()
			if err != nil {
				return nil, err
			}
			node.DiffType = MiguelKindUpdate
			node.Children = append(node.Children, child)
		}
	}

	return node, nil
}
