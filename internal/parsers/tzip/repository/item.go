package repository

import (
	"encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Item -
type Item struct {
	Network  string
	Address  string
	Metadata []byte
}

// Metadata -
type Metadata struct {
	tzip.TZIP
	Tokens struct {
		Static []struct {
			Name     string                 `json:"name"`
			Symbol   string                 `json:"symbol,omitempty"`
			Decimals *int64                 `json:"decimals,omitempty"`
			TokenID  int64                  `json:"token_id"`
			Extras   map[string]interface{} `json:"extras"`
		} `json:"static"`
	} `json:"tokens"`
}

// ToModel -
func (item Item) ToModel() (*Metadata, error) {
	t, err := time.Parse("2006 01 02 15 04", "2018 06 30 00 00")
	if err != nil {
		return nil, err
	}

	model := new(Metadata)
	model.Network = item.Network
	model.Address = item.Address
	model.Timestamp = t.UTC()
	model.OffChain = true

	err = json.Unmarshal(item.Metadata, &model)
	return model, err
}
