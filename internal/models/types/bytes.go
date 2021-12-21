package types

import (
	"encoding/hex"
	stdJSON "encoding/json"
	"errors"
)

// Bytes -
type Bytes stdJSON.RawMessage

// MarshalJSON returns b as the JSON encoding of b.
func (b Bytes) MarshalJSON() ([]byte, error) {
	if b == nil {
		return []byte("null"), nil
	}
	return b, nil
}

// UnmarshalJSON sets *b to a copy of data.
func (b *Bytes) UnmarshalJSON(data []byte) error {
	if b == nil {
		return errors.New("json.RawMessage: UnmarshalJSON on nil pointer")
	}
	*b = append((*b)[0:0], data...)
	return nil
}

// MustNewBytes -
func MustNewBytes(str string) Bytes {
	raw, _ := hex.DecodeString(str)
	return Bytes(raw)
}
