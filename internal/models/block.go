package models

import (
	"time"

	"github.com/tidwall/gjson"
)

// Block -
type Block struct {
	ID string `json:"-"`

	Network     string    `json:"network"`
	Hash        string    `json:"hash"`
	Level       int64     `json:"level"`
	Predecessor string    `json:"predecessor"`
	ChainID     string    `json:"chain_id"`
	Protocol    string    `json:"protocol"`
	Timestamp   time.Time `json:"timestamp"`
}

// GetID -
func (b *Block) GetID() string {
	return b.ID
}

// GetIndex -
func (b *Block) GetIndex() string {
	return "block"
}

// GetQueue -
func (b *Block) GetQueue() string {
	return ""
}

// Marshal -
func (b *Block) Marshal() ([]byte, error) {
	return nil, nil
}

// ParseElasticJSON -
func (b *Block) ParseElasticJSON(hit gjson.Result) {
	b.ID = hit.Get("_id").String()
	b.Network = hit.Get("_source.network").String()
	b.Protocol = hit.Get("_source.protocol").String()
	b.Level = hit.Get("_source.level").Int()
	b.Timestamp = hit.Get("_source.timestamp").Time()
	b.ChainID = hit.Get("_source.chain_id").String()
	b.Predecessor = hit.Get("_source.predecessor").String()
	b.Hash = hit.Get("_source.hash").String()
}
