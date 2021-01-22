package bcdast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
)

// List -
type List struct {
	Default

	Type AstNode

	Data []AstNode
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
			s.WriteString(strings.Repeat(indent, list.depth))
			s.WriteByte('{')
			s.WriteByte('\n')
			s.WriteString(strings.Repeat(indent, list.depth+1))
			s.WriteString(list.Data[i].String())
			s.WriteString(strings.Repeat(indent, list.depth))
			s.WriteByte('}')
			s.WriteByte('\n')
		}
	} else {
		s.WriteString(strings.Repeat(indent, list.depth))
		s.WriteString(list.Type.String())
	}
	return s.String()
}

// ParseType -
func (list *List) ParseType(untyped Untyped, id *int) error {
	if err := list.Default.ParseType(untyped, id); err != nil {
		return err
	}

	typ, err := typingNode(untyped.Args[0], list.depth, id)
	if err != nil {
		return err
	}
	list.Type = typ

	return nil
}

// ParseValue -
func (list *List) ParseValue(untyped Untyped) error {
	if untyped.Prim != PrimArray {
		return errors.Wrap(ErrInvalidPrim, "List.ParseValue")
	}

	list.Data = make([]AstNode, 0)

	for i := range untyped.Args {
		item, err := createByType(list.Type)
		if err != nil {
			return err
		}
		if err := item.ParseValue(untyped.Args[i]); err != nil {
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
