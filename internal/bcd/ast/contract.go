package ast

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

// Contract -
type Contract struct {
	Address

	Type Node
}

// NewContract -
func NewContract(depth int) *Contract {
	return &Contract{
		Address: Address{
			Default: NewDefault(consts.CONTRACT, 1, depth),
		},
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
	s.WriteString(strings.Repeat(base.DefaultIndent, c.depth))
	s.WriteString(c.Type.String())

	return s.String()
}

// ParseType -
func (c *Contract) ParseType(node *base.Node, id *int) error {
	if err := c.Default.ParseType(node, id); err != nil {
		return err
	}
	contractType, err := typingNode(node.Args[0], c.depth, id)
	if err != nil {
		return err
	}
	c.Type = contractType
	return nil
}

// ToJSONSchema -
func (c *Contract) ToJSONSchema() (*JSONSchema, error) {
	s := &JSONSchema{
		Prim:    c.Prim,
		Type:    JSONSchemaTypeString,
		Default: "",
	}
	// TODO: set tags
	// tags, err := kinds.CheckParameterForTags(nm.Parameter)
	// if err != nil {
	// 	return nil, err
	// }
	// if len(tags) == 1 {
	// 	s.Tag = tags[0]
	// }
	return s, nil
}
