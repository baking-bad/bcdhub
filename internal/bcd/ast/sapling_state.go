package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

// SaplingState -
type SaplingState struct {
	Default

	Type     Int
	MemoSize int64
}

// NewSaplingState -
func NewSaplingState(depth int) *SaplingState {
	return &SaplingState{
		Default: NewDefault(consts.SAPLINGSTATE, 1, depth),
	}
}

// MarshalJSON -
func (ss *SaplingState) MarshalJSON() ([]byte, error) {
	var builder bytes.Buffer
	builder.WriteString(`{"prim": "sapling_state", "args":[`)
	builder.WriteString(fmt.Sprintf(`{"int": "%d"}`, ss.MemoSize))
	builder.WriteByte(']')
	if len(ss.annots) > 0 {
		if _, err := builder.WriteString(fmt.Sprintf(`, "annots": ["%s"]`, strings.Join(ss.annots, `","`))); err != nil {
			return nil, err
		}
	}
	builder.WriteByte('}')
	return builder.Bytes(), nil
}

// String -
func (ss *SaplingState) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("[%d] %s memo_size=%d\n", ss.id, ss.Prim, ss.MemoSize))
	if ss.Type.Value != nil {
		s.WriteString(strings.Repeat(base.DefaultIndent, ss.depth))
		s.WriteString(fmt.Sprintf("Int=%d", ss.Type.Value))
	}
	return s.String()
}

// ParseType -
func (ss *SaplingState) ParseType(node *base.Node, id *int) error {
	if err := ss.Default.ParseType(node, id); err != nil {
		return err
	}

	if err := ss.Type.ParseType(node.Args[0], id); err != nil {
		return err
	}
	ss.MemoSize = node.Args[0].IntValue.Int64()

	return nil
}

// ParseValue -
func (ss *SaplingState) ParseValue(node *base.Node) error {
	return ss.Type.ParseValue(node)
}

// ToMiguel -
func (ss *SaplingState) ToMiguel() (*MiguelNode, error) {
	node, err := ss.Default.ToMiguel()
	if err != nil {
		return nil, err
	}
	node.Children = make([]*MiguelNode, 0)
	child, err := ss.Type.ToMiguel()
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, child)
	return node, nil
}

// ToBaseNode -
func (ss *SaplingState) ToBaseNode(optimized bool) (*base.Node, error) {
	return ss.Type.ToBaseNode(optimized)
}
