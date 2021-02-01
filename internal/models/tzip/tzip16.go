package tzip

import (
	"encoding/json"
)

// TZIP16 -
type TZIP16 struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	Version     string   `json:"version,omitempty"`
	License     *License `json:"license,omitempty"`
	Homepage    string   `json:"homepage,omitempty"`
	Authors     []string `json:"authors,omitempty"`
	Interfaces  []string `json:"interfaces,omitempty"`
	Views       []View   `json:"views,omitempty"`
}

// License -
type License struct {
	Name    string `json:"name"`
	Details string `json:"details,omitempty"`
}

// UnmarshalJSON -
func (license *License) UnmarshalJSON(data []byte) error {
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
