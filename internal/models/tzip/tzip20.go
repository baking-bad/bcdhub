package tzip

import "encoding/json"

// TZIP20 -
type TZIP20 struct {
	Events []Event `json:"events,omitempty"`
}

// Event -
type Event struct {
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	Pure            string                `json:"pure"`
	Implementations []EventImplementation `json:"implementations"`
}

// EventImplementation -
type EventImplementation struct {
	MichelsonParameterEvent       MichelsonParameterEvent `json:"michelson-parameter-event"`
	MichelsonInitialStorageEvent  Sections                `json:"michelson-initial-storage-event"`
	MichelsonExtendedStorageEvent Sections                `json:"michelson-extended-storage-event"`
}

// MichelsonParameterEvent -
type MichelsonParameterEvent struct {
	Sections
	Entrypoints []string `json:"entrypoints"`
}

// Sections -
type Sections struct {
	Parameter  json.RawMessage `json:"parameter"`
	ReturnType json.RawMessage `json:"return-type"`
	Code       json.RawMessage `json:"code"`
}
