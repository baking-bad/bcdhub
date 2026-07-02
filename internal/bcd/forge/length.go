package forge

import (
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
)

type length struct {
	Value uint32
}

// Unforge -
func (l *length) Unforge(data []byte) (uint32, error) {
	if len(data) < 4 {
		return 0, errors.Wrap(ErrTooFewBytes, fmt.Sprintf("Length.Unforge: %d < 4", len(data)))
	}

	l.Value = binary.BigEndian.Uint32(data[:4])
	return 4, nil
}

// Forge -
func (l *length) Forge() ([]byte, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, l.Value)
	return data, nil
}
