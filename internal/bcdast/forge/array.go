package forge

import (
	"fmt"

	"github.com/pkg/errors"
)

// Array -
type Array struct {
	Args []Node
}

// Unforge -
func (a *Array) Unforge(data []byte) (int, error) {
	l := length{}
	n, err := l.Unforge(data)
	if err != nil {
		return n, err
	}

	a.Args = make([]Node, 0)

	if l.Value == 0 {
		return n, nil
	}
	data = data[n:]

	if len(data) < l.Value {
		return 4, errors.Wrap(ErrTooFewBytes, fmt.Sprintf("String.Unforge: %d < %d", len(data), l.Value))
	}

	var count int
	for count < l.Value {
		// TODO: realize
	}

	return n + l.Value, nil
}
