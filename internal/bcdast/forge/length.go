package forge

import (
	"encoding/binary"
	"fmt"

	"github.com/pkg/errors"
)

type length struct {
	Value int
}

// Unforge -
func (l *length) Unforge(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, errors.Wrap(ErrTooFewBytes, fmt.Sprintf("Length.Unforge: %d < 4", len(data)))
	}

	l.Value = int(binary.BigEndian.Uint32(data[:4]))
	return 4, nil
}

// Forge -
func (l *length) Forge() ([]byte, error) {
	data := make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(l.Value))
	return data, nil
}
