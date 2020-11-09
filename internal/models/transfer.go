package models

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models/utils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// Transfer -
type Transfer struct {
	ID             string    `json:"-"`
	IndexedTime    int64     `json:"indexed_time"`
	Network        string    `json:"network"`
	Contract       string    `json:"contract"`
	Alias          string    `json:"alias,omitempty"`
	Initiator      string    `json:"initiator"`
	InitiatorAlias string    `json:"initiator_alias,omitempty"`
	Hash           string    `json:"hash"`
	Status         string    `json:"status"`
	Timestamp      time.Time `json:"timestamp"`
	Level          int64     `json:"level"`
	From           string    `json:"from"`
	FromAlias      string    `json:"from_alias,omitempty"`
	To             string    `json:"to"`
	ToAlias        string    `json:"to_alias,omitempty"`
	TokenID        int64     `json:"token_id"`
	Amount         float64   `json:"amount"`
	Counter        int64     `json:"counter"`
	Nonce          *int64    `json:"nonce,omitempty"`
}

// GetID -
func (t *Transfer) GetID() string {
	return t.ID
}

// GetIndex -
func (t *Transfer) GetIndex() string {
	return "transfer"
}

// GetQueues -
func (t *Transfer) GetQueues() []string {
	return []string{"transfers"}
}

// MarshalToQueue -
func (t *Transfer) MarshalToQueue() ([]byte, error) {
	return []byte(t.ID), nil
}

// LogFields -
func (t *Transfer) LogFields() logrus.Fields {
	return logrus.Fields{
		"network":  t.Network,
		"contract": t.Contract,
		"block":    t.Level,
		"from":     t.From,
		"to":       t.To,
	}
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
	return utils.GetFoundBy(keys, categories)
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
		Counter:     o.Counter,
		Nonce:       o.Nonce,
	}
}
