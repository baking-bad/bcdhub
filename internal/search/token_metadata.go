package search

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
)

// TokenResponse -
type TokenResponse struct{}

// Token -
type Token struct {
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
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return Item{
		Type:       t.GetIndex(),
		Value:      t.Address,
		Body:       t,
		Highlights: highlight,
		Network:    t.Network,
	}, nil
}
