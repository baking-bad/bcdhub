package types

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
)

// Bytes -
type Bytes json.RawMessage

// Scan scan value into Jsonb, implements sql.Scanner interface
func (b *Bytes) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	*b = bytes
	return nil
}

// Value return json value, implement driver.Valuer interface
func (b Bytes) Value() (driver.Value, error) {
	if len(b) == 0 {
		return nil, nil
	}
	return []byte(b), nil
}

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
