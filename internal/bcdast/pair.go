package bcdast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
)

// Pair -
type Pair struct {
	Default
	Args []AstNode
}

// NewPair -
func NewPair(depth int) *Pair {
	return &Pair{
		Default: NewDefault(consts.PAIR, -1, depth),
	}
}

// String -
func (p *Pair) String() string {
	var s strings.Builder
	s.WriteString(p.Default.String())
	for i := range p.Args {
		s.WriteString(strings.Repeat(indent, p.depth))
		s.WriteString(p.Args[i].String())
	}
	return s.String()
}

// MarshalJSON -
func (p *Pair) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.PAIR, p.annots, p.Args...)
}

// ParseType -
func (p *Pair) ParseType(untyped Untyped, id *int) error {
	if err := p.Default.ParseType(untyped, id); err != nil {
		return err
	}

	p.Args = make([]AstNode, 0)
	if len(untyped.Args) == 2 {
		for _, arg := range untyped.Args {
			child, err := typingNode(arg, p.depth, id)
			if err != nil {
				return err
			}
			p.Args = append(p.Args, child)
		}
	} else if len(untyped.Args) > 2 {
		child, err := typingNode(untyped.Args[0], p.depth, id)
		if err != nil {
			return err
		}
		p.Args = append(p.Args, child)

		newUntyped := Untyped{
			Prim: consts.PAIR,
			Args: untyped.Args[1:],
		}
		pairChild, err := typingNode(newUntyped, p.depth+1, id)
		if err != nil {
			return err
		}
		p.Args = append(p.Args, pairChild)
	}

	return nil
}

// ParseValue -
func (p *Pair) ParseValue(untyped Untyped) error {
	switch {
	case len(untyped.Args) == 2:
		for i := range p.Args {
			if err := p.Args[i].ParseValue(untyped.Args[i]); err != nil {
				return err
			}
		}
		return nil
	case len(untyped.Args) > 2:
		if err := p.Args[0].ParseValue(untyped.Args[0]); err != nil {
			return err
		}

		newUntyped := Untyped{
			Prim: consts.PAIR,
			Args: untyped.Args[1:],
		}
		return p.Args[1].ParseValue(newUntyped)
	default:
		return errors.Wrap(ErrInvalidArgsCount, "Pair.ParseValue")
	}
}

// ToMiguel -
func (p *Pair) ToMiguel() (*MiguelNode, error) {
	node := &MiguelNode{
		Prim:     p.Prim,
		Type:     consts.TypeNamedTuple,
		Name:     p.GetName(),
		Children: make([]*MiguelNode, 0),
	}

	for i := range p.Args {
		child, err := p.Args[i].ToMiguel()
		if err != nil {
			return nil, err
		}

		if p.Prim == p.Args[i].GetPrim() {
			node.Children = append(node.Children, child.Children...)
		} else {
			node.Children = append(node.Children, child)
		}
	}

	return node, nil
}
