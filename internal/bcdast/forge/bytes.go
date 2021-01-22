package forge

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
)

// Bytes -
type Bytes Node

// Unforge -
func (b *Bytes) Unforge(data []byte) (int, error) {
	l := length{}
	n, err := l.Unforge(data)
	if err != nil {
		return n, err
	}

	data = data[n:]

	if len(data) < l.Value {
		return 4, errors.Wrap(ErrTooFewBytes, fmt.Sprintf("Bytes.Unforge: %d < %d", len(data), l.Value))
	}

	s := hex.EncodeToString(data[:l.Value])
	b.BytesValue = &s
	return n + l.Value, nil
}
