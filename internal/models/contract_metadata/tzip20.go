package contract_metadata

import (
	"bytes"
	stdJSON "encoding/json"

	"github.com/baking-bad/bcdhub/internal/helpers"
)

// TZIP20 -
type TZIP20 struct {
	Events Events `json:"events,omitempty" pg:",type:jsonb"`
}

// Events -
type Events []Event

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

var null = []byte("null")

// Empty -
func (s Sections) Empty() bool {
	return bytes.HasSuffix(s.Code, null) && bytes.HasSuffix(s.Parameter, null) && bytes.HasSuffix(s.ReturnType, null)
}

// IsParameterEmpty -
func (s Sections) IsParameterEmpty() bool {
	return s.Parameter == nil || bytes.HasSuffix(s.Parameter, null)
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
