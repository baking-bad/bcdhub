package models

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/models/utils"
	"github.com/tidwall/gjson"
)

// TZIP -
type TZIP struct {
	Level   int64  `json:"level"`
	Address string `json:"address"`
	Network string `json:"network"`

	tzip.TZIP12
	tzip.TZIP16
}

// HasToken -
func (t TZIP) HasToken(network, address string, tokenID int64) bool {
	for i := range t.Tokens.Static {
		if t.Address == address && t.Network == network && t.Tokens.Static[i].TokenID == tokenID {
			return true
		}
	}
	return false
}

// ParseElasticJSON -
func (t *TZIP) ParseElasticJSON(resp gjson.Result) {
	t.Level = resp.Get("_source.level").Int()
	t.Address = resp.Get("_source.address").String()
	t.Network = resp.Get("_source.network").String()

	t.TZIP12.ParseElasticJSON(resp)
}

// GetID -
func (t *TZIP) GetID() string {
	return fmt.Sprintf("%s_%s", t.Network, t.Address)
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
	return []string{
		"tokens.name^8",
		"tokens.symbol^8",
		"address^7",
	}
}

// FoundByName -
func (t *TZIP) FoundByName(hit gjson.Result) string {
	keys := hit.Get("highlight").Map()
	categories := t.GetScores("")
	return utils.GetFoundBy(keys, categories)
}
