package forge

import (
	"fmt"

	"github.com/pkg/errors"
)

// Array -
type Array Node

// NewArray -
func NewArray() *Array {
	return &Array{
		Args: make([]*Node, 0),
		Prim: PrimArray,
	}
}

func newArrayFromNodes(nodes []*Node) *Array {
	return &Array{
		Args: nodes,
		Prim: PrimArray,
	}
}

// Unforge -
func (a *Array) Unforge(data []byte) (int, error) {
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

	if len(data) < l.Value {
		return 4, errors.Wrap(ErrTooFewBytes, fmt.Sprintf("String.Unforge: %d < %d", len(data), l.Value))
	}

	var count int
	for count < l.Value {
		unforger := NewMichelson()
		n, err := unforger.Unforge(data)
		if err != nil {
			return n, err
		}
		count += n
		data = data[n:]
		a.Args = append(a.Args, unforger.Nodes...)
	}

	return 4 + l.Value, nil
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
	l.Value = len(args)
	data, err := l.Forge()
	if err != nil {
		return nil, err
	}
	data = append(data, args...)

	return append([]byte{ByteArray}, data...), nil
}
