package search

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// SearchTypes
const (
	MetadataSearchType = "metadata"
)

// Metadata -
type Metadata struct{}

// GetIndex -
func (m Metadata) GetIndex() string {
	return models.DocTZIP
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
	var metadata tzip.TZIP
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}
	return models.Item{
		Type:       MetadataSearchType,
		Value:      metadata.Address,
		Body:       metadata,
		Highlights: highlight,
		Network:    metadata.Network,
	}, nil
}
