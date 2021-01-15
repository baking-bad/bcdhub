package search

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
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
	return models.DocTokenMetadata
}

// GetScores -
func (t Token) GetScores(search string) []string {
	return []string{
		"name^8",
		"symbol^8",
		"contract^7",
	}
}

// GetFields -
func (t Token) GetFields() []string {
	return []string{
		"name",
		"symbol",
		"contract",
	}
}

// Parse  -
func (t Token) Parse(highlight map[string][]string, data []byte) (interface{}, error) {
	var token tokenmetadata.TokenMetadata
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return models.Item{
		Type:       token.GetIndex(),
		Value:      token.Contract,
		Body:       token,
		Highlights: highlight,
		Network:    token.Network,
	}, nil
}
