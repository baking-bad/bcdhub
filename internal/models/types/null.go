package types

import (
	"database/sql/driver"
)

// NullString represents a string that may be null.
// NullString implements the Scanner interface so
// it can be used as a scan destination
type NullString struct {
	Str   string
	Valid bool // Valid is true if Str is not NULL
}

// NewNullString -
func NewNullString(val *string) NullString {
	if val == nil {
		return NullString{"", false}
	}

	return NullString{*val, true}
}

// UnmarshalJSON -
func (ns *NullString) UnmarshalJSON(data []byte) error {
	if len(data) < 2 {
		ns.Valid = false
		return nil
	}

	ns.Valid = true
	ns.Str = string(data)[0 : len(data)-1]
	return nil
}

// MarshalJSON -
func (ns NullString) MarshalJSON() ([]byte, error) {
	if ns.Valid {
		return []byte(ns.Str), nil
	}
	return nil, nil
}

// Scan implements the Scanner interface.
func (ns *NullString) Scan(value interface{}) error {
	ns.Str, ns.Valid = "", false

	if value == nil {
		return nil
	}

	switch val := value.(type) {
	case string:
		ns.Str, ns.Valid = val, true
	case []byte:
		ns.Str, ns.Valid = string(val), true
	}

	return nil
}

// Value implements the driver Valuer interface.
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.Str, nil
}

// String -
func (ns NullString) String() string {
	if !ns.Valid {
		return ""
	}
	return ns.Str
}

// EqualString -
func (ns NullString) EqualString(value string) bool {
	if !ns.Valid {
		return false
	}
	return value == ns.Str
}

// Set -
func (ns *NullString) Set(value interface{}) error {
	ns.Str, ns.Valid = "", false

	if value == nil {
		return nil
	}

	if val, ok := value.(string); ok {
		ns.Str, ns.Valid = val, true
	}
	return nil
}
