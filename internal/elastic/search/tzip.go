package search

import (
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Metadata -
type Metadata struct{}

// GetIndex -
func (m Metadata) GetIndex() string {
	return "tzip"
}

// GetScores -
func (m Metadata) GetScores(search string) []string {
	return []string{
		"name^10",
		"authors^10",
		"description^8",
		"homepage^4",
	}
}

// GetFields -
func (m Metadata) GetFields() []string {
	return []string{
		"name",
		"homepage",
		"description",
		"authors",
	}
}

// Parse  -
func (m Metadata) Parse(highlight map[string][]string, data []byte) (interface{}, error) {
	var token tzip.TZIP
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return Item{
		Type:       "metadata",
		Value:      token.Address,
		Body:       token,
		Highlights: highlight,
		Network:    token.Network,
	}, nil
}
