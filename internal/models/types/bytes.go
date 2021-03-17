package types

import (
	"database/sql/driver"
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
