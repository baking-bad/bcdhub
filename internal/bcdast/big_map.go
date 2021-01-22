package bcdast

import (
	"fmt"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/pkg/errors"
)

// BigMap -
type BigMap struct {
	Default

	KeyType   AstNode
	ValueType AstNode

	Data map[AstNode]AstNode
	Ptr  *int64
}

// NewBigMap -
func NewBigMap(depth int) *BigMap {
	return &BigMap{
		Default: NewDefault(consts.BIGMAP, 2, depth),
		Data:    make(map[AstNode]AstNode),
	}
}

// String -
func (m *BigMap) String() string {
	var s strings.Builder

	s.WriteString(m.Default.String())
	switch {
	case m.Ptr != nil:
		s.WriteString(strings.Repeat(indent, m.depth))
		s.WriteString(fmt.Sprintf("Ptr=%d\n", *m.Ptr))
	case len(m.Data) > 0:
		for key, val := range m.Data {
			s.WriteString(strings.Repeat(indent, m.depth))
			s.WriteByte('{')
			s.WriteByte('\n')
			s.WriteString(strings.Repeat(indent, m.depth+1))
			s.WriteString(key.String())
			s.WriteString(strings.Repeat(indent, m.depth+1))
			s.WriteString(val.String())
			s.WriteString(strings.Repeat(indent, m.depth))
			s.WriteByte('}')
			s.WriteByte('\n')
		}
	default:
		s.WriteString(strings.Repeat(indent, m.depth))
		s.WriteString(m.KeyType.String())
		s.WriteString(strings.Repeat(indent, m.depth))
		s.WriteString(m.ValueType.String())
	}

	return s.String()
}

// MarshalJSON -
func (m *BigMap) MarshalJSON() ([]byte, error) {
	return marshalJSON(consts.BIGMAP, m.annots, m.KeyType, m.ValueType)
}

// ParseType -
func (m *BigMap) ParseType(untyped Untyped, id *int) error {
	if err := m.Default.ParseType(untyped, id); err != nil {
		return err
	}

	keyType, err := typingNode(untyped.Args[0], m.depth, id)
	if err != nil {
		return err
	}
	m.KeyType = keyType

	valType, err := typingNode(untyped.Args[1], m.depth, id)
	if err != nil {
		return err
	}
	m.ValueType = valType

	return nil
}

// ParseValue -
func (m *BigMap) ParseValue(untyped Untyped) error {
	switch {
	case untyped.IntValue != nil:
		m.Ptr = untyped.IntValue
	case untyped.Prim == PrimArray:
		data, err := createMapFromElts(untyped.Args, m.KeyType, m.ValueType)
		if err != nil {
			return err
		}
		m.Data = data
	default:
		return errors.Wrap(ErrInvalidPrim, fmt.Sprintf("BigMap.ParseValue (%s)", untyped.Prim))
	}
	return nil
}

// ToMiguel -
func (m *BigMap) ToMiguel() (*MiguelNode, error) {
	node, err := m.Default.ToMiguel()
	if err != nil {
		return nil, err
	}

	switch {
	case m.Ptr != nil:
		node.Value = *m.Ptr
		return node, nil
	default:
		node.Children = make([]*MiguelNode, 0)
		for key, value := range m.Data {
			keyChild, err := key.ToMiguel()
			if err != nil {
				return nil, err
			}
			if keyChild != nil {
				child, err := value.ToMiguel()
				if err != nil {
					return nil, err
				}

				name, err := getMapKeyName(keyChild)
				if err != nil {
					return nil, err
				}
				child.Name = name
				node.Children = append(node.Children, child)
			}
		}

		return node, nil
	}

}
