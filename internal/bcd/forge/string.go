package forge

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/pkg/errors"
)

// String -
type String base.Node

// Unforge -
func (s *String) Unforge(data []byte) (uint32, error) {
	l := new(length)
	n, err := l.Unforge(data)
	if err != nil {
		return n, err
	}

	data = data[n:]

	if uint32(len(data)) < l.Value { // #nosec G115 -- unforged data is bounded by protocol operation size limits, never close to uint32 max
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
	l.Value = uint32(len(b)) // #nosec G115 -- forged string is bounded by protocol operation size limits, never close to uint32 max
	data, err := l.Forge()
	if err != nil {
		return nil, err
	}

	data = append(data, b...)

	return append([]byte{ByteString}, data...), nil
}
