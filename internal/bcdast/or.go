package bcdast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
)

// Or -
type Or struct {
	Default
	Args []AstNode

	Left  AstNode
	Right AstNode
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
		s.WriteString(strings.Repeat(indent, or.depth))
		s.WriteString(or.Left.String())
		s.WriteByte(' ')
	case or.Right != nil:
		s.WriteString(consts.Right)
		s.WriteByte(' ')
		s.WriteString(or.Default.String())
		s.WriteString(strings.Repeat(indent, or.depth))
		s.WriteString(or.Right.String())
		s.WriteByte(' ')
	default:
		s.WriteString(or.Default.String())
		for i := range or.Args {
			s.WriteString(strings.Repeat(indent, or.depth))
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
func (or *Or) ParseType(untyped Untyped, id *int) error {
	if err := or.Default.ParseType(untyped, id); err != nil {
		return err
	}

	if len(untyped.Args) > 2 || len(untyped.Args) == 0 {
		return errors.Wrap(ErrInvalidArgsCount, "Or.ParseValue")
	}

	or.Args = make([]AstNode, 0, len(untyped.Args))
	for _, arg := range untyped.Args {
		child, err := typingNode(arg, or.depth, id)
		if err != nil {
			return err
		}
		or.Args = append(or.Args, child)
	}

	return nil
}

// ParseValue -
func (or *Or) ParseValue(untyped Untyped) error {
	if len(untyped.Args) > 2 || len(untyped.Args) == 0 {
		return errors.Wrap(ErrInvalidArgsCount, "Or.ParseValue")
	}

	switch untyped.Prim {
	case consts.Left:
		if err := or.Args[0].ParseValue(untyped.Args[0]); err != nil {
			return err
		}
		or.Left = or.Args[0]
	case consts.Right:
		if err := or.Args[1].ParseValue(untyped.Args[0]); err != nil {
			return err
		}
		or.Right = or.Args[1]
	default:
		return errors.Wrap(ErrInvalidPrim, "Or.ParseValue")
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
