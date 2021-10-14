package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

// Constant -
type Constant struct {
	Default

	KeyHash string
}

// NewConstant -
func NewConstant(depth int) *Constant {
	return &Constant{
		Default: NewDefault(consts.CONSTANT, 1, depth),
	}
}

// MarshalJSON -
func (c *Constant) MarshalJSON() ([]byte, error) {
	var builder bytes.Buffer
	builder.WriteString(`{"prim": "constant", "args":[`)
	builder.WriteString(fmt.Sprintf(`{"string": "%s"}`, c.KeyHash))
	builder.WriteByte(']')
	builder.WriteByte('}')
	return builder.Bytes(), nil
}

// String -
func (c *Constant) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("[%d] %s key_hash=%s", c.ID, c.Prim, c.KeyHash))
	return s.String()
}

// ParseType -
func (c *Constant) ParseType(node *base.Node, id *int) error {
	if err := c.Default.ParseType(node, id); err != nil {
		return err
	}

	if len(node.Args) == 0 {
		return nil
	}

	if node.Args[0].StringValue == nil {
		return nil
	}

	c.KeyHash = *node.Args[0].StringValue
	return nil
}

// ParseValue -
func (c *Constant) ParseValue(node *base.Node) error {
	return nil
}

// ToMiguel -
func (c *Constant) ToMiguel() (*MiguelNode, error) {
	return c.Default.ToMiguel()
}

// Docs -
func (c *Constant) Docs(inferredName string) ([]Typedef, string, error) {
	return []Typedef{
		{
			Name: c.GetName(),
			Type: fmt.Sprintf("%s(%s)", c.Prim, c.KeyHash),
		},
	}, c.Prim, nil
}

// Distinguish -
func (c *Constant) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*Constant)
	if !ok {
		return nil, nil
	}
	return c.Default.Distinguish(&second.Default)
}

// EqualType -
func (c *Constant) EqualType(node Node) bool {
	return c.Default.EqualType(node)
}

// FindByName -
func (c *Constant) FindByName(name string, isEntrypoint bool) Node {
	if c.GetName() == name {
		return c
	}
	return nil
}
