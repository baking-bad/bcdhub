package models

import (
	"reflect"
	"time"

	"github.com/tidwall/gjson"
)

// TokenMetadata -
type TokenMetadata struct {
	ID string `json:"-"`

	Interface       string    `json:"interface"`
	Contract        string    `json:"contract"`
	RegistryAddress string    `json:"registry_address,omitempty"`
	Network         string    `json:"network"`
	Timestamp       time.Time `json:"timestamp"`
	Level           int64     `json:"level,omitempty"`

	TokenID  int64                  `json:"token_id"`
	Symbol   string                 `json:"symbol"`
	Name     string                 `json:"name"`
	Decimals int64                  `json:"decimals"`
	Extras   map[string]interface{} `json:"extras,omitempty"`
}

// ParseElasticJSON -
func (tm *TokenMetadata) ParseElasticJSON(resp gjson.Result) {
	tm.ID = resp.Get("_id").String()

	tm.Interface = resp.Get("_source.interface").String()
	tm.Contract = resp.Get("_source.contract").String()
	tm.RegistryAddress = resp.Get("_source.registry_address").String()
	tm.Network = resp.Get("_source.network").String()
	tm.Timestamp = resp.Get("_source.timestamp").Time().UTC()
	tm.Level = resp.Get("_source.level").Int()

	tm.TokenID = resp.Get("_source.token_id").Int()
	tm.Symbol = resp.Get("_source.symbol").String()
	tm.Name = resp.Get("_source.name").String()
	tm.Decimals = resp.Get("_source.decimals").Int()

	tm.Extras = make(map[string]interface{})
	for k, v := range resp.Get("_source.extras").Map() {
		tm.Extras[k] = v.Value()
	}
}

// GetID -
func (tm *TokenMetadata) GetID() string {
	return tm.ID
}

// GetIndex -
func (tm *TokenMetadata) GetIndex() string {
	return "token_metadata"
}

// GetQueue -
func (tm *TokenMetadata) GetQueue() string {
	return ""
}

// Marshal -
func (tm *TokenMetadata) Marshal() ([]byte, error) {
	return nil, nil
}

// GetScores -
func (tm *TokenMetadata) GetScores(search string) []string {
	return []string{
		"symbol^10",
		"name^10",
		"contract^7",
		"registry_address^6",
	}
}

// FoundByName -
func (tm *TokenMetadata) FoundByName(hit gjson.Result) string {
	keys := hit.Get("highlight").Map()
	categories := tm.GetScores("")
	return getFoundBy(keys, categories)
}

// Is - checks is object equals by key fields: `Contract`, `Network`, `TokenID`
func (tm TokenMetadata) Is(other TokenMetadata) bool {
	return tm.Contract == other.Contract && tm.Network == other.Network && tm.TokenID == other.TokenID
}

// Compare - full compare objects
func (tm TokenMetadata) Compare(other TokenMetadata) bool {
	return tm.Is(other) &&
		tm.RegistryAddress == other.RegistryAddress &&
		tm.Symbol == other.Symbol &&
		tm.Name == other.Name &&
		tm.Decimals == other.Decimals &&
		reflect.DeepEqual(tm.Extras, other.Extras)
}
