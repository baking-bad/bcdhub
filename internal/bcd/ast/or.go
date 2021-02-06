package ast

import (
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
		s.WriteString(strings.Repeat(base.DefaultIndent, or.depth))
		s.WriteString(or.Left.String())
		s.WriteByte(' ')
	case or.Right != nil:
		s.WriteString(consts.Right)
		s.WriteByte(' ')
		s.WriteString(or.Default.String())
		s.WriteString(strings.Repeat(base.DefaultIndent, or.depth))
		s.WriteString(or.Right.String())
		s.WriteByte(' ')
	default:
		s.WriteString(or.Default.String())
		for i := range or.Args {
			s.WriteString(strings.Repeat(base.DefaultIndent, or.depth))
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
		return errors.Wrap(base.ErrInvalidArgsCount, "Or.ParseValue")
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
		return errors.Wrap(base.ErrInvalidArgsCount, "Or.ParseValue")
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
		return errors.Wrap(base.ErrInvalidPrim, "Or.ParseValue")
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
				Title: key,
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
				return errors.Wrapf(base.ErrInvalidType, "Or.FromJSONSchema %T", val)
			}
			orMap = arrVal
			break
		}
	}
	schemaKey, ok := orMap["schemaKey"]
	if !ok {
		return errors.Wrap(base.ErrJSONDataIsAbsent, "Or.FromJSONSchema")
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
		return errors.Wrap(base.ErrJSONDataIsAbsent, "Or.FromJSONSchema")
	}

	return nil
}
