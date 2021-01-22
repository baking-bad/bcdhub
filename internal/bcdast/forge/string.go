package forge

import (
	"fmt"

	"github.com/pkg/errors"
)

// String -
type String Node

// Unforge -
func (s *String) Unforge(data []byte) (int, error) {
	l := length{}
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
