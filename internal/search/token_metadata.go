package search

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
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

// GetID -
func (t *Token) GetID() string {
	return t.ID
}

// GetIndex -
func (t *Token) GetIndex() string {
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
func (t Token) Parse(highlight map[string][]string, data []byte) (*Item, error) {
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &Item{
		Type:       t.GetIndex(),
		Value:      t.Contract,
		Body:       &t,
		Highlights: highlight,
		Network:    t.Network,
	}, nil
}

// Prepare -
func (t *Token) Prepare(model models.Model) {
	tm, ok := model.(*tokenmetadata.TokenMetadata)
	if !ok {
		return
	}

	t.ID = helpers.GenerateID()
	t.Contract = tm.Contract
	t.Decimals = tm.Decimals
	t.Extras = tm.Extras
	t.Level = tm.Level
	t.Name = tm.Name
	t.Network = tm.Network.String()
	t.Symbol = tm.Symbol
	t.Timestamp = tm.Timestamp
	t.TokenID = tm.TokenID
}
