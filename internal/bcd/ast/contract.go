package ast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/formatter"
	"github.com/tidwall/gjson"
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

// Docs -
func (c *Contract) Docs(inferredName string) ([]Typedef, string, error) {
	name := getNameDocString(c, inferredName)
	typedef := Typedef{
		Name: name,
		Type: fmt.Sprintf("contract(%s)", c.Type.GetPrim()),
		Args: make([]TypedefArg, 0),
	}
	if !isSimpleDocType(c.Type.GetPrim()) {
		b, err := json.Marshal(c.Type)
		if err != nil {
			return nil, "", err
		}
		paramName := fmt.Sprintf("%s_param", c.GetName())
		parameter, err := formatter.MichelineToMichelson(gjson.ParseBytes(b), true, formatter.DefLineSize)
		if err != nil {
			return nil, "", err
		}

		typedef.Type = fmt.Sprintf("contract(%s)", makeVarDocString(paramName))
		paramTypedef := Typedef{
			Name: paramName,
			Type: parameter,
		}
		return []Typedef{typedef, paramTypedef}, typedef.Type, nil
	}

	return nil, typedef.Type, nil
}
