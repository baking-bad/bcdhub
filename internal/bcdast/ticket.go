package bcdast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

// Ticket -
type Ticket struct {
	Default

	Type   AstNode
	Paired AstNode

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
	s.WriteString(strings.Repeat(indent, t.depth))
	s.WriteString(t.Type.String())
	return s.String()
}

// MarshalJSON -
func (t *Ticket) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.TICKET, t.annots, t.Type)
}

// ParseType -
func (t *Ticket) ParseType(untyped Untyped, id *int) error {
	if err := t.Default.ParseType(untyped, id); err != nil {
		return err
	}

	typ, err := typingNode(untyped.Args[0], t.depth, id)
	if err != nil {
		return err
	}
	t.Type = typ
	t.Paired = &Pair{
		Default: NewDefault(consts.PAIR, -1, t.depth+1),
		Args: []AstNode{
			NewAddress(t.depth + 2),
			&Pair{
				Default: NewDefault(consts.PAIR, -1, t.depth+2),
				Args: []AstNode{
					typ,
					NewNat(t.depth + 3),
				},
			},
		},
	}

	return nil
}

// ParseValue -
func (t *Ticket) ParseValue(untyped Untyped) error {
	return t.Type.ParseValue(untyped)
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
