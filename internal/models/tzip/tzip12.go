package tzip

import (
	"reflect"

	"github.com/tidwall/gjson"
)

// TZIP12 -
type TZIP12 struct {
	Tokens TokenMetadataType `json:"tokens,omitempty"`
}

// TokenMetadataType -
type TokenMetadataType struct {
	Static []TokenMetadata `json:"static,omitempty"`
	// Dynamic []TokenMetadata `json:"dynamic,omitempty"`
}

// TokenMetadata -
type TokenMetadata struct {
	RegistryAddress string                 `json:"registry_address"`
	TokenID         int64                  `json:"token_id"`
	Symbol          string                 `json:"symbol"`
	Name            string                 `json:"name"`
	Decimals        int64                  `json:"decimals"`
	Extras          map[string]interface{} `json:"extras"`
}

// ParseElasticJSON -
func (t *TZIP12) ParseElasticJSON(resp gjson.Result) {
	tokensJSON := resp.Get("_source.tokens.static")
	if tokensJSON.Exists() {
		t.Tokens = TokenMetadataType{
			Static: make([]TokenMetadata, 0),
		}

		for _, item := range tokensJSON.Array() {
			extras := make(map[string]interface{})
			for key, value := range item.Get("extras").Map() {
				extras[key] = value.Value()
			}

			t.Tokens.Static = append(t.Tokens.Static, TokenMetadata{
				Symbol:          item.Get("symbol").String(),
				Decimals:        item.Get("decimals").Int(),
				TokenID:         item.Get("token_id").Int(),
				Name:            item.Get("name").String(),
				RegistryAddress: item.Get("registry_address").String(),
				Extras:          extras,
			})
		}
	}
}

// Compare - full compare objects
func (tm TokenMetadata) Compare(other TokenMetadata) bool {
	return tm.RegistryAddress == other.RegistryAddress &&
		tm.Symbol == other.Symbol &&
		tm.Name == other.Name &&
		tm.TokenID == other.TokenID &&
		tm.Decimals == other.Decimals &&
		reflect.DeepEqual(tm.Extras, other.Extras)
}
