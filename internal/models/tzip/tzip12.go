package tzip

import (
	"reflect"
)

// TZIP12 -
type TZIP12 struct {
	Tokens *TokenMetadataType `json:"tokens,omitempty"`
}

// TokenMetadataType -
type TokenMetadataType struct {
	Static []TokenMetadataEntity `json:"static,omitempty"`
	// Dynamic []TokenMetadata `json:"dynamic,omitempty"`
}

// TokenMetadataEntity -
type TokenMetadataEntity struct {
	RegistryAddress string                 `json:"registry_address"`
	TokenID         int64                  `json:"token_id"`
	Symbol          string                 `json:"symbol"`
	Name            string                 `json:"name"`
	Decimals        *int64                 `json:"decimals,omitempty"`
	Extras          map[string]interface{} `json:"extras"`
}

// Compare - full compare objects
func (tm TokenMetadataEntity) Compare(other TokenMetadataEntity) bool {
	return tm.RegistryAddress == other.RegistryAddress &&
		tm.Symbol == other.Symbol &&
		tm.Name == other.Name &&
		tm.TokenID == other.TokenID &&
		*tm.Decimals == *other.Decimals &&
		reflect.DeepEqual(tm.Extras, other.Extras)
}
