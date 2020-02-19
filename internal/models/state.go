package models

import (
	"time"

	"github.com/tidwall/gjson"
)

// State -
type State struct {
	ID        string    `json:"-"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Network   string    `json:"network"`
	Type      string    `json:"type"`
}

// ParseElasticJSON -
func (s *State) ParseElasticJSON(hit gjson.Result) {
	s.ID = hit.Get("_id").String()
	s.Network = hit.Get("_source.type").String()
	s.Type = hit.Get("_source.network").String()
	s.Level = hit.Get("_source.level").Int()
	s.Timestamp = hit.Get("_source.timestamp").Time()
}
