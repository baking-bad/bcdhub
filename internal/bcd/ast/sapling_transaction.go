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

	Data     *Bytes
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
	if len(st.Annots) > 0 {
		if _, err := builder.WriteString(fmt.Sprintf(`, "annots": ["%s"]`, strings.Join(st.Annots, `","`))); err != nil {
			return nil, err
		}
	}
	builder.WriteByte('}')
	return builder.Bytes(), nil
}

// String -
func (st *SaplingTransaction) String() string {
	var s strings.Builder
	s.WriteString(fmt.Sprintf("[%d] %s memo_size=%d\n", st.ID, st.Prim, st.MemoSize))
	if st.Data.Value != nil {
		s.WriteString(strings.Repeat(consts.DefaultIndent, st.Depth))
		s.WriteString(fmt.Sprintf("Int=%d", st.Data.Value))
	}
	return s.String()
}

// ParseType -
func (st *SaplingTransaction) ParseType(node *base.Node, id *int) error {
	if err := st.Default.ParseType(node, id); err != nil {
		return err
	}

	st.Data = NewBytes(st.Depth)
	st.MemoSize = node.Args[0].IntValue.Int64()

	return nil
}

// ParseValue -
func (st *SaplingTransaction) ParseValue(node *base.Node) error {
	return st.Data.ParseValue(node)
}

// ToMiguel -
func (st *SaplingTransaction) ToMiguel() (*MiguelNode, error) {
	node, err := st.Default.ToMiguel()
	if err != nil {
		return nil, err
	}
	node.Value = st.Data.miguelValue()
	return node, nil
}

// ToBaseNode -
func (st *SaplingTransaction) ToBaseNode(optimized bool) (*base.Node, error) {
	return st.Data.ToBaseNode(optimized)
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

// Distinguish -
func (st *SaplingTransaction) Distinguish(x Distinguishable) (*MiguelNode, error) {
	second, ok := x.(*SaplingTransaction)
	if !ok {
		return nil, nil
	}

	node, err := st.ToMiguel()
	if err != nil {
		return nil, err
	}
	switch {
	case st.Data.Value == nil && second.Data.Value == nil:
	case st.Data.Value != nil && second.Data.Value == nil:
		node.setDiffType(MiguelKindCreate)
	case st.Data.Value == nil && second.Data.Value != nil:
		node.setDiffType(MiguelKindDelete)
	case st.Data.Value != nil && second.Data.Value != nil:
		node.From = second.Data.Value
		node.setDiffType(MiguelKindUpdate)
	}
	return node, nil
}

// ToParameters -
func (st *SaplingTransaction) ToParameters() ([]byte, error) {
	return st.Data.ToParameters()
}

// FromJSONSchema -
func (st *SaplingTransaction) FromJSONSchema(data map[string]interface{}) error {
	for key := range data {
		if key == st.GetName() {
			st.Data.Value = data[key]
			st.Data.ValueKind = valueKindBytes
		}
	}
	return nil
}
