package forge

import (
	"encoding/hex"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/pkg/errors"
)

// Michelson -
type Michelson struct {
	Nodes []*base.Node
}

// NewMichelson -
func NewMichelson() *Michelson {
	return &Michelson{
		Nodes: make([]*base.Node, 0),
	}
}

// UnforgeString -
func (m *Michelson) UnforgeString(s string) (int, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	return m.Unforge(b)
}

// Unforge -
func (m *Michelson) Unforge(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	switch b[0] {
	case ByteInt:
		unforger := NewInt()
		n, err = unforger.Unforge(b[1:])
		if err != nil {
			return n, err
		}
		m.Nodes = append(m.Nodes, (*base.Node)(unforger))
	case ByteString:
		unforger := new(String)
		n, err = unforger.Unforge(b[1:])
		if err != nil {
			return n, err
		}
		m.Nodes = append(m.Nodes, (*base.Node)(unforger))
	case ByteArray:
		unforger := NewArray()
		n, err = unforger.Unforge(b[1:])
		if err != nil {
			return n, err
		}
		m.Nodes = append(m.Nodes, (*base.Node)(unforger))
	case BytePrim, BytePrimAnnots, BytePrimArg, BytePrimArgAnnots, BytePrimArgs, BytePrimArgsAnnots:
		argsCount := int((b[0] - 3) / 2)
		hasAnnots := b[0]%2 == 0
		unforger := NewObject(argsCount, hasAnnots)
		n, err = unforger.Unforge(b[1:])
		if err != nil {
			return n, err
		}
		m.Nodes = append(m.Nodes, &base.Node{
			Prim:   unforger.Prim,
			Args:   unforger.Args,
			Annots: unforger.Annots,
		})
	case ByteGeneralPrim:
		unforger := NewObject(-1, true)
		n, err = unforger.Unforge(b[1:])
		if err != nil {
			return n, err
		}
		m.Nodes = append(m.Nodes, &base.Node{
			Prim:   unforger.Prim,
			Args:   unforger.Args,
			Annots: unforger.Annots,
		})
	case ByteBytes:
		unforger := new(Bytes)
		n, err = unforger.Unforge(b[1:])
		if err != nil {
			return n, err
		}
		m.Nodes = append(m.Nodes, (*base.Node)(unforger))
	default:
		return 1, errors.Wrap(ErrUnknownTypeCode, fmt.Sprintf("%x", b[0]))
	}
	return n + 1, nil
}

// Forge -
func (m *Michelson) Forge() ([]byte, error) {
	data := make([]byte, 0)
	for i := range m.Nodes {
		var forger Forger
		switch {
		case m.Nodes[i].IntValue != nil:
			forger = (*Int)(m.Nodes[i])
		case m.Nodes[i].StringValue != nil:
			forger = (*String)(m.Nodes[i])
		case m.Nodes[i].BytesValue != nil:
			forger = (*Bytes)(m.Nodes[i])
		case m.Nodes[i].Prim == PrimArray:
			forger = (*Array)(m.Nodes[i])
		case m.Nodes[i].Prim != "":
			forger = &Object{*m.Nodes[i], -1, false}
		}
		body, err := forger.Forge()
		if err != nil {
			return nil, err
		}
		data = append(data, body...)
	}
	return data, nil
}
