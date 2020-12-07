package search

import (
	"encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
)

// TokenResponse -
type TokenResponse struct {
	Name      string                 `json:"name"`
	Symbol    string                 `json:"symbol"`
	TokenID   int64                  `json:"token_id"`
	Network   string                 `json:"network"`
	Address   string                 `json:"address"`
	Level     int64                  `json:"level"`
	Timestamp time.Time              `json:"timestamp"`
	Decimals  *int64                 `json:"decimals,omitempty"`
	Extras    map[string]interface{} `json:"extras,omitempty"`
}

// Token -
type Token struct{}

// GetIndex -
func (t Token) GetIndex() string {
	return "tzip"
}

// GetScores -
func (t Token) GetScores(search string) []string {
	return []string{
		"tokens.static.name^8",
		"tokens.static.symbol^8",
		"address^7",
	}
}

// GetFields -
func (t Token) GetFields() []string {
	return []string{
		"tokens.static.name",
		"tokens.static.symbol",
		"address",
	}
}

// Parse  -
func (t Token) Parse(highlight map[string][]string, data []byte) (interface{}, error) {
	var token models.TZIP
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	if token.Tokens == nil {
		return nil, nil
	}
	items := make([]Item, len(token.Tokens.Static))
	for i := range token.Tokens.Static {
		items[i] = Item{
			Type:  t.GetIndex(),
			Value: token.Address,
			Body: TokenResponse{
				Network:   token.Network,
				Address:   token.Address,
				Level:     token.Level,
				Timestamp: token.Timestamp,
				Name:      token.Tokens.Static[i].Name,
				Symbol:    token.Tokens.Static[i].Symbol,
				TokenID:   token.Tokens.Static[i].TokenID,
				Decimals:  token.Tokens.Static[i].Decimals,
				Extras:    token.Tokens.Static[i].Extras,
			},
			Highlights: highlight,
			Network:    token.Network,
		}
	}
	return items, nil
}
