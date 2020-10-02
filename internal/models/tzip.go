package models

import "github.com/tidwall/gjson"

// TZIP -
type TZIP struct {
	ID string `json:"-"`

	TZIP16
}

// TZIP16 -
type TZIP16 struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Version     string      `json:"version"`
	License     string      `json:"license"`
	Homepage    string      `json:"homepage"`
	Authors     []string    `json:"authors"`
	Interfaces  []string    `json:"interfaces"`
	Views       interface{} `json:"views"`
}

// ParseElasticJSON -
func (t *TZIP) ParseElasticJSON(resp gjson.Result) {
	t.ID = resp.Get("_id").String()

	t.Name = resp.Get("_source.name").String()
	t.Description = resp.Get("_source.description").String()
	t.Version = resp.Get("_source.version").String()
	t.License = resp.Get("_source.license").Time().String()
	t.Homepage = resp.Get("_source.homepage").String()

	t.Authors = parseStringArray(resp, "_source.authors")
	t.Interfaces = parseStringArray(resp, "_source.interfaces")
	t.Views = resp.Get("_source.views").Value()
}

// GetID -
func (t *TZIP) GetID() string {
	return t.ID
}

// GetIndex -
func (t *TZIP) GetIndex() string {
	return "tzip"
}

// GetQueue -
func (t *TZIP) GetQueue() string {
	return ""
}

// Marshal -
func (t *TZIP) Marshal() ([]byte, error) {
	return nil, nil
}

// GetScores -
func (t *TZIP) GetScores(search string) []string {
	return []string{}
}

// FoundByName -
func (t *TZIP) FoundByName(hit gjson.Result) string {
	keys := hit.Get("highlight").Map()
	categories := t.GetScores("")
	return getFoundBy(keys, categories)
}
