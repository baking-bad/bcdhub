package types

import (
	"database/sql/driver"
	"encoding/json"
)

// JSONB -
type JSONB map[string]interface{}

// Value -
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return []byte(`{}`), nil
	}
	return json.Marshal(j)
}

// Scan -
func (j *JSONB) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &j)
}
