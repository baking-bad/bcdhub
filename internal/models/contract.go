package models

import (
	"time"
)

// Contract - entity for contract
type Contract struct {
	ID          string       `json:"-"`
	Network     string       `json:"network"`
	Level       int64        `json:"level"`
	Timestamp   time.Time    `json:"timestamp"`
	Balance     int64        `json:"balance"`
	Fingerprint *Fingerprint `json:"fingerprint,omitempty"`
	Language    string       `json:"language,omitempty"`

	Tags        []string `json:"tags,omitempty"`
	Hardcoded   []string `json:"hardcoded,omitempty"`
	FailStrings []string `json:"fail_strings,omitempty"`
	Primitives  []string `json:"primitives,omitempty"`
	Annotations []string `json:"annotations,omitempty"`
	Entrypoints []string `json:"entrypoints,omitempty"`

	Address  string `json:"address"`
	Manager  string `json:"manager,omitempty"`
	Delegate string `json:"delegate,omitempty"`

	ProjectID string `json:"project_id,omitempty"`
	FoundBy   string `json:"found_by,omitempty"`
	Group     *Group `json:"group,omitempty"`
}

// Fingerprint -
type Fingerprint struct {
	Code      string `json:"code"`
	Storage   string `json:"storage"`
	Parameter string `json:"parameter"`
}

// Group -
type Group struct {
	Count int64    `json:"count"`
	Top   []string `json:"top"`
}
