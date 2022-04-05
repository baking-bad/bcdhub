package search

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// TokenResponse -
type TokenResponse struct{}

// Token -
type Token struct {
	ID        string                 `json:"-"`
	Name      string                 `json:"name"`
	Symbol    string                 `json:"symbol"`
	TokenID   uint64                 `json:"token_id"`
	Network   string                 `json:"network"`
	Contract  string                 `json:"contract"`
	Level     int64                  `json:"level"`
	Timestamp time.Time              `json:"timestamp"`
	Decimals  *int64                 `json:"decimals,omitempty"`
	Extras    map[string]interface{} `json:"extras,omitempty"`
}

// NewToken -
func NewToken(network types.Network, model *tokenmetadata.TokenMetadata) Token {
	var t Token
	t.ID = helpers.GenerateID()
	t.Contract = model.Contract
	t.Decimals = model.Decimals
	t.Extras = model.Extras
	t.Level = model.Level
	t.Name = model.Name
	t.Network = network.String()
	t.Symbol = model.Symbol
	t.Timestamp = model.Timestamp.UTC()
	t.TokenID = model.TokenID
	return t
}

// GetID -
func (t Token) GetID() string {
	return t.ID
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
	}
}

// GetFields -
func (t Token) GetFields() []string {
	return []string{
		"name",
		"symbol",
	}
}

// Parse  -
func (t Token) Parse(highlight map[string][]string, data []byte) (*Item, error) {
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &Item{
		Type:       t.GetIndex(),
		Value:      fmt.Sprintf("token %d in %s", t.TokenID, t.Contract),
		Body:       &t,
		Highlights: highlight,
		Network:    t.Network,
	}, nil
}
