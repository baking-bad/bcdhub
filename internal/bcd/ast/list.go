package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
)

// List -
type List struct {
	Default

	Type Node

	Data []Node
}

// NewList -
func NewList(depth int) *List {
	return &List{
		Default: NewDefault(consts.LIST, -1, depth),
	}
}

// MarshalJSON -
func (list *List) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.LIST, list.annots, list.Type)
}

// String -
func (list *List) String() string {
	var s strings.Builder

	s.WriteString(list.Default.String())
	if len(list.Data) > 0 {
		for i := range list.Data {
			s.WriteString(strings.Repeat(base.DefaultIndent, list.depth))
			s.WriteByte('{')
			s.WriteByte('\n')
			s.WriteString(strings.Repeat(base.DefaultIndent, list.depth+1))
			s.WriteString(list.Data[i].String())
			s.WriteString(strings.Repeat(base.DefaultIndent, list.depth))
			s.WriteByte('}')
			s.WriteByte('\n')
		}
	} else {
		s.WriteString(strings.Repeat(base.DefaultIndent, list.depth))
		s.WriteString(list.Type.String())
	}
	return s.String()
}

// ParseType -
func (list *List) ParseType(node *base.Node, id *int) error {
	if err := list.Default.ParseType(node, id); err != nil {
		return err
	}

	typ, err := typingNode(node.Args[0], list.depth, id)
	if err != nil {
		return err
	}
	list.Type = typ

	return nil
}

// ParseValue -
func (list *List) ParseValue(node *base.Node) error {
	if node.Prim != base.PrimArray {
		return errors.Wrap(base.ErrInvalidPrim, "List.ParseValue")
	}

	list.Data = make([]Node, 0)

	for i := range node.Args {
		item, err := createByType(list.Type)
		if err != nil {
			return err
		}
		if err := item.ParseValue(node.Args[i]); err != nil {
			return err
		}
		list.Data = append(list.Data, item)
	}

	return nil
}

// ToMiguel -
func (list *List) ToMiguel() (*MiguelNode, error) {
	node, err := list.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	node.Children = make([]*MiguelNode, 0)
	for i := range list.Data {
		child, err := list.Data[i].ToMiguel()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	return node, nil
}

// ToBaseNode -
func (list *List) ToBaseNode(optimized bool) (*base.Node, error) {
	return arrayToBaseNode(list.Data, optimized)
}

// ToJSONSchema -
func (list *List) ToJSONSchema() (*JSONSchema, error) {
	s := &JSONSchema{
		Prim:    list.Prim,
		Type:    JSONSchemaTypeArray,
		Title:   list.GetName(),
		Default: make([]interface{}, 0),
		Items: &SchemaKey{
			Type:       JSONSchemaTypeObject,
			Required:   make([]string, 0),
			Properties: make(map[string]*JSONSchema),
		},
	}

	if err := setChildSchema(list.Type, true, s); err != nil {
		return nil, err
	}

	return wrapObject(s), nil
}
