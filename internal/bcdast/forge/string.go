package forge

import (
	"fmt"

	"github.com/pkg/errors"
)

// String -
type String Node

// Unforge -
func (s *String) Unforge(data []byte) (int, error) {
	l := new(length)
	n, err := l.Unforge(data)
	if err != nil {
		return n, err
	}

	data = data[n:]

	if len(data) < l.Value {
		return 4, errors.Wrap(ErrTooFewBytes, fmt.Sprintf("String.Unforge: %d < %d", len(data), l.Value))
	}

	str := string(data[:l.Value])
	s.StringValue = &str

	return n + l.Value, nil
}

// Forge -
func (s *String) Forge() ([]byte, error) {
	l := new(length)
	b := []byte(*s.StringValue)
	l.Value = len(b)
	data, err := l.Forge()
	if err != nil {
		return nil, err
	}

	data = append(data, b...)

	return append([]byte{ByteString}, data...), nil
}
