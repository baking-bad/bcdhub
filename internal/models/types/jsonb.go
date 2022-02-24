package types

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"

	"github.com/pkg/errors"
)

// JSONB -
type JSONB map[string]interface{}

// Value -
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return []byte(`{}`), nil
	}
	data, err := json.Marshal(j)
	if err != nil {
		return data, err
	}
	return j.escape(data), nil
}

// Scan -
func (j *JSONB) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if !ok {
		return errors.Errorf("invalid type of JSONB data: %T", data)
	}
	return json.Unmarshal(data, &j)
}

func (j *JSONB) escape(data []byte) []byte {
	return bytes.ReplaceAll(data, []byte("\\u0000"), []byte{})
}
