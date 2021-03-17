package search

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
)

// Contract -
type Contract struct {
	Network   string    `json:"network"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Language  string    `json:"language,omitempty"`

	Hash        string   `json:"hash"`
	Tags        []string `json:"tags,omitempty"`
	Hardcoded   []string `json:"hardcoded,omitempty"`
	FailStrings []string `json:"fail_strings,omitempty"`
	Annotations []string `json:"annotations,omitempty"`
	Entrypoints []string `json:"entrypoints,omitempty"`

	Address  string `json:"address"`
	Manager  string `json:"manager,omitempty"`
	Delegate string `json:"delegate,omitempty"`

	TxCount       int64     `json:"tx_count"`
	LastAction    time.Time `json:"last_action"`
	FoundBy       string    `json:"found_by,omitempty"`
	Alias         string    `json:"alias,omitempty"`
	DelegateAlias string    `json:"delegate_alias,omitempty"`
}

// GetIndex -
func (c Contract) GetIndex() string {
	return models.DocContracts
}

// GetScores -
func (c Contract) GetScores(search string) []string {
	if helpers.IsAddress(search) {
		return []string{
			"address^10",
			"alias^9",
			"tags^9",
			"entrypoints^8",
			"fail_strings^6",
			"language^4",
			"annotations^3",
			"delegate^2",
			"hardcoded^2",
			"manager",
		}
	}
	return []string{
		"alias^20",
		"tags^9",
		"entrypoints^8",
		"fail_strings^6",
		"language^4",
		"annotations^3",
		"delegate^2",
		"hardcoded^2",
		"manager",
		"address",
	}
}

// GetFields -
func (c Contract) GetFields() []string {
	return []string{
		"address",
		"alias",
		"tags",
		"entrypoints",
		"fail_strings",
		"language",
		"annotations",
		"delegate",
		"hardcoded",
		"manager",
	}
}

// Parse  -
func (c Contract) Parse(highlight map[string][]string, data []byte) (interface{}, error) {
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return Item{
		Type:       c.GetIndex(),
		Value:      c.Address,
		Body:       c,
		Highlights: highlight,
		Network:    c.Network,
	}, nil
}
