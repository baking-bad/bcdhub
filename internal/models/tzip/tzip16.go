package tzip

import (
	"database/sql/driver"
	"errors"
	"fmt"

	jsoniter "github.com/json-iterator/go"

	"github.com/lib/pq"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// TZIP16 -
type TZIP16 struct {
	Name        string         `json:"name,omitempty"`
	Description string         `json:"description,omitempty"`
	Version     string         `json:"version,omitempty"`
	License     *License       `json:"license,omitempty" gorm:"type:jsonb"`
	Homepage    string         `json:"homepage,omitempty"`
	Authors     pq.StringArray `json:"authors,omitempty" gorm:"type:text[]"`
	Interfaces  pq.StringArray `json:"interfaces,omitempty" gorm:"type:text[]"`
	Views       Views          `json:"views,omitempty" gorm:"type:jsonb"`
}

// License -
type License struct {
	Name    string `json:"name"`
	Details string `json:"details,omitempty"`
}

// UnmarshalJSON -
func (license *License) UnmarshalJSON(data []byte) error {
	if len(data) <= 2 {
		return nil
	}
	switch data[0] {
	case '"':
		if err := json.Unmarshal(data, &license.Name); err != nil {
			return err
		}
	case '{':
		var buf struct {
			Name    string `json:"name"`
			Details string `json:"details,omitempty"`
		}
		if err := json.Unmarshal(data, &buf); err != nil {
			return err
		}
		license.Name = buf.Name
		license.Details = buf.Details
	}
	return nil
}

// IsEmpty -
func (license *License) IsEmpty() bool {
	return license.Name == "" && license.Details == ""
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (license *License) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, license)
}

// Value return json value, implement driver.Valuer interface
func (license *License) Value() (driver.Value, error) {
	if license == nil {
		return []byte(`{}`), nil
	}
	return json.Marshal(license)
}

// Views -
type Views []View

// Scan scan value into Jsonb, implements sql.Scanner interface
func (views *Views) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, views)
}

// Value return json value, implement driver.Valuer interface
func (views Views) Value() (driver.Value, error) {
	if views == nil {
		return []byte(`[]`), nil
	}
	return json.Marshal(views)
}

// View -
type View struct {
	Name            string               `json:"name"`
	Description     string               `json:"description"`
	Implementations []ViewImplementation `json:"implementations"`
}

// ViewImplementation -
type ViewImplementation struct {
	MichelsonStorageView Sections `json:"michelsonStorageView"`
}
