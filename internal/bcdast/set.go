package bcdast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
)

// Set -
type Set struct {
	Default

	Type AstNode
	Data []AstNode
}

// NewSet -
func NewSet(depth int) *Set {
	return &Set{
		Default: NewDefault(consts.SET, -1, depth),
	}
}

// String -
func (set *Set) String() string {
	var s strings.Builder

	s.WriteString(set.Default.String())
	if len(set.Data) > 0 {
		for i := range set.Data {
			s.WriteString(strings.Repeat(indent, set.depth))
			s.WriteByte('{')
			s.WriteByte('\n')
			s.WriteString(strings.Repeat(indent, set.depth+1))
			s.WriteString(set.Data[i].String())
			s.WriteString(strings.Repeat(indent, set.depth))
			s.WriteByte('}')
			s.WriteByte('\n')
		}
	} else {
		s.WriteString(strings.Repeat(indent, set.depth))
		s.WriteString(set.Type.String())
	}
	return s.String()
}

// MarshalJSON -
func (set *Set) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.SET, set.annots, set.Type)
}

// ParseType -
func (set *Set) ParseType(untyped Untyped, id *int) error {
	if err := set.Default.ParseType(untyped, id); err != nil {
		return err
	}

	typ, err := typingNode(untyped.Args[0], set.depth, id)
	if err != nil {
		return err
	}
	set.Type = typ

	return nil
}

// ParseValue -
func (set *Set) ParseValue(untyped Untyped) error {
	if untyped.Prim != PrimArray {
		return errors.Wrap(ErrInvalidPrim, "List.ParseValue")
	}

	set.Data = make([]AstNode, 0)

	for i := range untyped.Args {
		item, err := createByType(set.Type)
		if err != nil {
			return err
		}
		if err := item.ParseValue(untyped.Args[i]); err != nil {
			return err
		}
		set.Data = append(set.Data, item)
	}

	return nil
}

// ToMiguel -
func (set *Set) ToMiguel() (*MiguelNode, error) {
	node, err := set.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	node.Children = make([]*MiguelNode, 0)
	for i := range set.Data {
		child, err := set.Data[i].ToMiguel()
		if err != nil {
			return nil, err
		}
		node.Children = append(node.Children, child)
	}

	return node, nil
}
