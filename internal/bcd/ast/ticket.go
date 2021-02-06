package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

//  TODO: pack/unpack

// Ticket -
type Ticket struct {
	Default

	Type   Node
	Paired Node

	depth int
}

// NewTicket -
func NewTicket(depth int) *Ticket {
	return &Ticket{
		depth: depth,
		Default: Default{
			Prim:      consts.TICKET,
			argsCount: 1,
		},
	}
}

// String -
func (t *Ticket) String() string {
	var s strings.Builder
	s.WriteString(t.Default.String())
	s.WriteString(strings.Repeat(base.DefaultIndent, t.depth))
	s.WriteString(t.Type.String())
	return s.String()
}

// MarshalJSON -
func (t *Ticket) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.TICKET, t.annots, t.Type)
}

// ParseType -
func (t *Ticket) ParseType(node *base.Node, id *int) error {
	if err := t.Default.ParseType(node, id); err != nil {
		return err
	}

	typ, err := typingNode(node.Args[0], t.depth, id)
	if err != nil {
		return err
	}
	t.Type = typ
	t.Paired = &Pair{
		Default: NewDefault(consts.PAIR, -1, t.depth+1),
		Args: []Node{
			NewAddress(t.depth + 2),
			&Pair{
				Default: NewDefault(consts.PAIR, -1, t.depth+2),
				Args: []Node{
					typ,
					NewNat(t.depth + 3),
				},
			},
		},
	}

	return nil
}

// ParseValue -
func (t *Ticket) ParseValue(node *base.Node) error {
	return t.Type.ParseValue(node)
}

// ToMiguel -
func (t *Ticket) ToMiguel() (*MiguelNode, error) {
	node, err := t.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	node.Children = make([]*MiguelNode, 0)
	child, err := t.Paired.ToMiguel()
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, child)
	return node, nil
}

// ToBaseNode -
func (t *Ticket) ToBaseNode(optimized bool) (*base.Node, error) {
	return t.Paired.ToBaseNode(optimized)
}

// ToParameters -
func (t *Ticket) ToParameters() ([]byte, error) {
	return t.Paired.ToParameters()
}
