package models

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/tidwall/gjson"
)

// Transfer -
type Transfer struct {
	ID          string    `json:"-"`
	IndexedTime int64     `json:"indexed_time"`
	Network     string    `json:"network"`
	Contract    string    `json:"contract"`
	Initiator   string    `json:"initiator"`
	Alias       string    `json:"alias,omitempty"`
	Hash        string    `json:"hash"`
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	Level       int64     `json:"level"`
	From        string    `json:"from"`
	FromAlias   string    `json:"from_alias,omitempty"`
	To          string    `json:"to"`
	ToAlias     string    `json:"to_alias,omitempty"`
	TokenID     int64     `json:"token_id"`
	Amount      float64   `json:"amount"`
	Nonce       *int64    `json:"nonce"`
	Counter     *int64    `json:"counter"`
}

// ParseElasticJSON -
func (t *Transfer) ParseElasticJSON(resp gjson.Result) {
	t.ID = resp.Get("_id").String()
	t.IndexedTime = resp.Get("_source.indexed_time").Int()
	t.Network = resp.Get("_source.network").String()
	t.Contract = resp.Get("_source.contract").String()
	t.Initiator = resp.Get("_source.initiator").String()
	t.Alias = resp.Get("_source.alias").String()
	t.Hash = resp.Get("_source.hash").String()
	t.Status = resp.Get("_source.status").String()
	t.Timestamp = resp.Get("_source.timestamp").Time().UTC()
	t.Level = resp.Get("_source.level").Int()
	t.From = resp.Get("_source.from").String()
	t.FromAlias = resp.Get("_source.from_alias").String()
	t.To = resp.Get("_source.to").String()
	t.ToAlias = resp.Get("_source.to_alias").String()
	t.TokenID = resp.Get("_source.token_id").Int()
	t.Amount = resp.Get("_source.amount").Float()
	nonce := resp.Get("_source.indexed_time").Int()
	t.Nonce = &nonce
	counter := resp.Get("_source.indexed_time").Int()
	t.Counter = &counter
}

// GetID -
func (t *Transfer) GetID() string {
	return t.ID
}

// GetIndex -
func (t *Transfer) GetIndex() string {
	return "transfer"
}

// GetQueue -
func (t *Transfer) GetQueue() string {
	return "transfers"
}

// Marshal -
func (t *Transfer) Marshal() ([]byte, error) {
	return []byte(t.ID), nil
}

// GetScores -
func (t *Transfer) GetScores(search string) []string {
	return []string{
		"contract^8",
		"hash^7",
		"from^7",
		"to^6",
		"initiator",
	}
}

// FoundByName -
func (t *Transfer) FoundByName(hit gjson.Result) string {
	keys := hit.Get("highlight").Map()
	categories := t.GetScores("")
	return getFoundBy(keys, categories)
}

// EmptyTransfer -
func EmptyTransfer(o Operation) *Transfer {
	return &Transfer{
		ID:          helpers.GenerateID(),
		IndexedTime: o.IndexedTime,
		Network:     o.Network,
		Contract:    o.Destination,
		Hash:        o.Hash,
		Status:      o.Status,
		Timestamp:   o.Timestamp,
		Level:       o.Level,
		Initiator:   o.Source,
	}
}
