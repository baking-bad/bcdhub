package ast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

// Ticket -
type Ticket struct {
	Default

	Type       Node
	PairedType Node

	depth int
}

// NewTicket -
func NewTicket(depth int) *Ticket {
	return &Ticket{
		depth: depth,
		Default: Default{
			Prim:      consts.TICKET,
			ArgsCount: 1,
		},
	}
}

// String -
func (t *Ticket) String() string {
	var s strings.Builder
	s.WriteString(t.Default.String())
	s.WriteString(strings.Repeat(consts.DefaultIndent, t.depth))
	s.WriteString(t.Type.String())
	return s.String()
}

// MarshalJSON -
func (t *Ticket) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.TICKET, t.Annots, t.Type)
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

	pair := NewPair(t.depth)
	(*id)++
	pair.ID = *id

	address := NewAddress(t.depth + 2)
	(*id)++
	address.ID = *id

	internalPair := NewPair(t.depth)
	(*id)++
	internalPair.ID = *id

	nat := NewNat(t.depth)
	(*id)++
	nat.ID = *id

	internalPair.Args = []Node{
		Copy(typ),
		nat,
	}
	pair.Args = []Node{
		address,
		internalPair,
	}
	t.PairedType = pair

	return nil
}

// ParseValue -
func (t *Ticket) ParseValue(node *base.Node) error {
	return t.PairedType.ParseValue(node)
}

// ToMiguel -
func (t *Ticket) ToMiguel() (*MiguelNode, error) {
	node, err := t.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	node.Children = make([]*MiguelNode, 0)
	child, err := t.PairedType.ToMiguel()
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, child)
	return node, nil
}

// ToBaseNode -
func (t *Ticket) ToBaseNode(optimized bool) (*base.Node, error) {
	return t.PairedType.ToBaseNode(optimized)
}

// ToParameters -
func (t *Ticket) ToParameters() ([]byte, error) {
	return t.PairedType.ToParameters()
}

// Docs -
func (t *Ticket) Docs(inferredName string) ([]Typedef, string, error) {
	name := getNameDocString(t, inferredName)
	docs, varName, err := t.Type.Docs(name)
	if err != nil {
		return nil, "", err
	}

	optName := fmt.Sprintf("ticket(%s)", varName)
	if isSimpleDocType(docs[0].Type) {
		return nil, optName, nil
	}
	return docs, optName, nil
}

// Distinguish -
func (t *Ticket) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Ticket)
	if !ok {
		return nil, nil
	}
	return t.PairedType.Distinguish(second.PairedType)
}

// FromJSONSchema -
func (t *Ticket) FromJSONSchema(data map[string]interface{}) error {
	return nil
}

// EqualType -
func (t *Ticket) EqualType(node Node) bool {
	if !t.Default.EqualType(node) {
		return false
	}
	second, ok := node.(*Ticket)
	if !ok {
		return false
	}

	return t.Type.EqualType(second.Type)
}

// FindByName -
func (t *Ticket) FindByName(name string, isEntrypoint bool) Node {
	if t.GetName() == name {
		return t
	}
	return nil
}
