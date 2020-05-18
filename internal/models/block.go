package models

import (
	"time"

	"github.com/tidwall/gjson"
)

// Block -
type Block struct {
	ID string `json:"-"`

	Network     string    `json:"network" example:"mainnet"`
	Hash        string    `json:"hash" example:"BLyAEwaXShJuZasvUezHUfLqzZ48V8XrPvXF2wRaH15tmzEpsHT"`
	Level       int64     `json:"level" example:"24"`
	Predecessor string    `json:"predecessor" example:"BMWVEwEYw9m5iaHzqxDfkPzZTV4rhkSouRh3DkVMVGkxZ3EVaNs"`
	ChainID     string    `json:"chain_id" example:"NetXdQprcVkpaWU"`
	Protocol    string    `json:"protocol" example:"PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY"`
	Timestamp   time.Time `json:"timestamp" example:"2018-06-30T18:05:27Z"`
}

// GetID -
func (b *Block) GetID() string {
	return b.ID
}

// GetIndex -
func (b *Block) GetIndex() string {
	return "block"
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
