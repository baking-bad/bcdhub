package ast

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

// SaplingTransaction -
type SaplingTransaction struct {
	Default

	Type     Int
	MemoSize int64
}

// NewSaplingTransaction -
func NewSaplingTransaction(depth int) *SaplingTransaction {
	return &SaplingTransaction{
		Default: NewDefault(consts.SAPLINGTRANSACTION, 1, depth),
	}
}

// MarshalJSON -
func (st *SaplingTransaction) MarshalJSON() ([]byte, error) {
	var builder bytes.Buffer
	builder.WriteString(`{"prim": "sapling_transaction", "args":[`)
	builder.WriteString(fmt.Sprintf(`{"int": "%d"}`, st.MemoSize))
	builder.WriteByte(']')
	if len(st.annots) > 0 {
		if _, err := builder.WriteString(fmt.Sprintf(`, "annots": ["%s"]`, strings.Join(st.annots, `","`))); err != nil {
			return nil, err
		}
	}
	builder.WriteByte('}')
	return builder.Bytes(), nil
}

// String -
func (st *SaplingTransaction) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("[%d] %s memo_size=%d\n", st.id, st.Prim, st.MemoSize))
	if st.Type.Value != nil {
		s.WriteString(strings.Repeat(consts.DefaultIndent, st.depth))
		s.WriteString(fmt.Sprintf("Int=%d", st.Type.Value))
	}
	return s.String()
}

// ParseType -
func (st *SaplingTransaction) ParseType(node *base.Node, id *int) error {
	if err := st.Default.ParseType(node, id); err != nil {
		return err
	}

	if err := st.Type.ParseType(node.Args[0], id); err != nil {
		return err
	}
	st.MemoSize = node.Args[0].IntValue.Int64()

	return nil
}

// ParseValue -
func (st *SaplingTransaction) ParseValue(node *base.Node) error {
	return st.Type.ParseValue(node)
}

// ToMiguel -
func (st *SaplingTransaction) ToMiguel() (*MiguelNode, error) {
	node, err := st.Default.ToMiguel()
	if err != nil {
		return nil, err
	}
	node.Children = make([]*MiguelNode, 0)
	child, err := st.Type.ToMiguel()
	if err != nil {
		return nil, err
	}
	node.Children = append(node.Children, child)
	return node, nil
}

// ToBaseNode -
func (st *SaplingTransaction) ToBaseNode(optimized bool) (*base.Node, error) {
	return st.Type.ToBaseNode(optimized)
}

// Docs -
func (st *SaplingTransaction) Docs(inferredName string) ([]Typedef, string, error) {
	typ := fmt.Sprintf("%s(%d)", st.Prim, st.MemoSize)
	return []Typedef{
		{
			Name: st.GetName(),
			Type: typ,
		},
	}, typ, nil
}
