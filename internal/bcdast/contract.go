package bcdast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

// Contract -
type Contract struct {
	Default

	Type AstNode
}

// NewContract -
func NewContract(depth int) *Contract {
	return &Contract{
		Default: NewDefault(consts.CONTRACT, 1, depth),
	}
}

// MarshalJSON -
func (c *Contract) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.CONTRACT, c.annots, c.Type)
}

// String -
func (c *Contract) String() string {
	var s strings.Builder

	s.WriteString(c.Default.String())
	s.WriteString(strings.Repeat(indent, c.depth))
	s.WriteString(c.Type.String())

	return s.String()
}

// ParseType -
func (c *Contract) ParseType(untyped Untyped, id *int) error {
	if err := c.Default.ParseType(untyped, id); err != nil {
		return err
	}
	contractType, err := typingNode(untyped.Args[0], c.depth, id)
	if err != nil {
		return err
	}
	c.Type = contractType
	return nil
}