package forge

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/pkg/errors"
)

// Array -
type Array base.Node

// NewArray -
func NewArray() *Array {
	return &Array{
		Args: make([]*base.Node, 0),
		Prim: PrimArray,
	}
}

func newArrayFromNodes(nodes []*base.Node) *Array {
	return &Array{
		Args: nodes,
		Prim: PrimArray,
	}
}

// Unforge -
func (a *Array) Unforge(data []byte) (uint32, error) {
	a.Prim = PrimArray

	l := new(length)
	n, err := l.Unforge(data)
	if err != nil {
		return n, err
	}

	if l.Value == 0 {
		return n, nil
	}
	data = data[n:]

	if uint32(len(data)) < l.Value { // #nosec G115 -- unforged data is bounded by protocol operation size limits, never close to uint32 max
		return 4, errors.Wrap(ErrTooFewBytes, fmt.Sprintf("String.Unforge: %d < %d", len(data), l.Value))
	}

	data = data[:l.Value]
	for len(data) > 0 {
		unforger := NewMichelson()
		consumed, err := unforger.Unforge(data)
		if err != nil {
			return consumed, err
		}
		if consumed == 0 || consumed > uint32(len(data)) { // #nosec G115 -- len(data) <= l.Value, a uint32
			return n, errors.Wrap(ErrTooFewBytes, fmt.Sprintf("Array.Unforge: element overruns sequence (%d > %d)", consumed, len(data)))
		}
		data = data[consumed:]
		a.Args = append(a.Args, unforger.Nodes...)
	}

	return n + l.Value, nil
}

// Forge -
func (a *Array) Forge() ([]byte, error) {
	forger := NewMichelson()
	forger.Nodes = a.Args
	args, err := forger.Forge()
	if err != nil {
		return nil, err
	}
	l := new(length)
	l.Value = uint32(len(args)) // #nosec G115 -- forged args are bounded by protocol operation size limits, never close to uint32 max
	data, err := l.Forge()
	if err != nil {
		return nil, err
	}
	data = append(data, args...)

	return append([]byte{ByteArray}, data...), nil
}
