package models

import "github.com/tidwall/gjson"

// Metadata -
type Metadata struct {
	ID        string            `json:"-"`
	Parameter map[string]string `json:"parameter"`
	Storage   map[string]string `json:"storage"`
}

// ParseElasticJSON -
func (m *Metadata) ParseElasticJSON(hit gjson.Result) {
	m.ID = hit.Get("_id").String()
	m.Parameter = map[string]string{}
	for k, v := range hit.Get("_source.parameter").Map() {
		m.Parameter[k] = v.String()
	}

	m.Storage = map[string]string{}
	for k, v := range hit.Get("_source.storage").Map() {
		m.Storage[k] = v.String()
	}
}

// GetID -
func (m *Metadata) GetID() string {
	return m.ID
}

// GetIndex -
func (m *Metadata) GetIndex() string {
	return "metadata"
}

// GetQueue -
func (m *Metadata) GetQueue() string {
	return ""
}

// Marshal -
func (m *Metadata) Marshal() ([]byte, error) {
	return nil, nil
}
