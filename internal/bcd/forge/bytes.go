package forge

import (
	"encoding/hex"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/bcd/base"
	"github.com/pkg/errors"
)

// Bytes -
type Bytes base.Node

// Unforge -
func (b *Bytes) Unforge(data []byte) (uint32, error) {
	l := new(length)
	n, err := l.Unforge(data)
	if err != nil {
		return n, err
	}

	data = data[n:]

	if uint32(len(data)) < l.Value { // #nosec G115 -- unforged data is bounded by protocol operation size limits, never close to uint32 max
		return 4, errors.Wrap(ErrTooFewBytes, fmt.Sprintf("Bytes.Unforge: %d < %d", len(data), l.Value))
	}

	s := hex.EncodeToString(data[:l.Value])
	b.BytesValue = &s
	return n + l.Value, nil
}

// Forge -
func (b *Bytes) Forge() ([]byte, error) {
	body, err := hex.DecodeString(*b.BytesValue)
	if err != nil {
		return nil, err
	}

	l := new(length)
	l.Value = uint32(len(body)) // #nosec G115 -- forged bytes are bounded by protocol operation size limits, never close to uint32 max
	data, err := l.Forge()
	if err != nil {
		return nil, err
	}

	data = append(data, body...)
	return append([]byte{ByteBytes}, data...), nil
}
