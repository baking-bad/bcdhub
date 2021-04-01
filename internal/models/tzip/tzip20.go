package tzip

import (
	"database/sql/driver"
	stdJSON "encoding/json"
	"errors"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/helpers"
)

// TZIP20 -
type TZIP20 struct {
	Events Events `json:"events,omitempty" gorm:"type:jsonb"`
}

// Events -
type Events []Event

// Scan scan value into Jsonb, implements sql.Scanner interface
func (events *Events) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	return json.Unmarshal(bytes, events)
}

// Value return json value, implement driver.Valuer interface
func (events Events) Value() (driver.Value, error) {
	if events == nil {
		return []byte(`[]`), nil
	}
	return json.Marshal(events)
}

// Event -
type Event struct {
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Implementations []EventImplementation `json:"implementations"`
}

// EventImplementation -
type EventImplementation struct {
	MichelsonParameterEvent       *MichelsonParameterEvent       `json:"michelsonParameterEvent,omitempty"`
	MichelsonInitialStorageEvent  *MichelsonInitialStorageEvent  `json:"michelsonInitialStorageEvent,omitempty"`
	MichelsonExtendedStorageEvent *MichelsonExtendedStorageEvent `json:"michelsonExtendedStorageEvent,omitempty"`
}

// MichelsonParameterEvent -
type MichelsonParameterEvent struct {
	Sections
	Entrypoints []string `json:"entrypoints,omitempty"`
}

// InEntrypoints -
func (event MichelsonParameterEvent) InEntrypoints(entrypoint string) bool {
	return helpers.StringInArray(entrypoint, event.Entrypoints)
}

// Is -
func (event MichelsonParameterEvent) Is(entrypoint string) bool {
	return !event.Empty() && event.InEntrypoints(entrypoint)
}

// Sections -
type Sections struct {
	Parameter  stdJSON.RawMessage `json:"parameter"`
	ReturnType stdJSON.RawMessage `json:"returnType"`
	Code       stdJSON.RawMessage `json:"code"`
}

var null = " null"

// Empty -
func (s Sections) Empty() bool {
	return string(s.Code) == null && string(s.Parameter) == null && string(s.ReturnType) == null
}

// IsParameterEmpty -
func (s Sections) IsParameterEmpty() bool {
	return string(s.Parameter) == null
}

// MichelsonInitialStorageEvent -
type MichelsonInitialStorageEvent struct {
	Sections
}

// MichelsonExtendedStorageEvent -
type MichelsonExtendedStorageEvent struct {
	Sections
	Entrypoints []string `json:"entrypoints,omitempty"`
}

// InEntrypoints -
func (event MichelsonExtendedStorageEvent) InEntrypoints(entrypoint string) bool {
	return helpers.StringInArray(entrypoint, event.Entrypoints)
}

// Is -
func (event MichelsonExtendedStorageEvent) Is(entrypoint string) bool {
	return !event.Empty() && event.InEntrypoints(entrypoint)
}
