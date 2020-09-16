package models

import (
	"time"

	"github.com/tidwall/gjson"
)

// Migration -
type Migration struct {
	ID          string `json:"-"`
	IndexedTime int64  `json:"indexed_time"`

	Network      string    `json:"network"`
	Protocol     string    `json:"protocol"`
	PrevProtocol string    `json:"prev_protocol,omitempty"`
	Hash         string    `json:"hash,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	Level        int64     `json:"level"`
	Address      string    `json:"address"`
	Kind         string    `json:"kind"`
}

// GetID -
func (m *Migration) GetID() string {
	return m.ID
}

// GetIndex -
func (m *Migration) GetIndex() string {
	return "migration"
}

// GetQueue -
func (m *Migration) GetQueue() string {
	return "migrations"
}

// Marshal -
func (m *Migration) Marshal() ([]byte, error) {
	return []byte(m.ID), nil
}

// ParseElasticJSON -
func (m *Migration) ParseElasticJSON(resp gjson.Result) {
	m.ID = resp.Get("_id").String()
	m.IndexedTime = resp.Get("_source.indexed_time").Int()

	m.Protocol = resp.Get("_source.protocol").String()
	m.PrevProtocol = resp.Get("_source.prev_protocol").String()
	m.Hash = resp.Get("_source.hash").String()
	m.Network = resp.Get("_source.network").String()
	m.Timestamp = resp.Get("_source.timestamp").Time().UTC()
	m.Level = resp.Get("_source.level").Int()
	m.Address = resp.Get("_source.address").String()
	m.Kind = resp.Get("_source.kind").String()
}
