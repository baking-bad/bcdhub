package search

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
)

// Metadata -
type Metadata struct {
	Level       int64     `json:"level,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty"`
	Address     string    `json:"address"`
	Network     string    `json:"network"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Homepage    string    `json:"homepage,omitempty"`
	Authors     []string  `json:"authors,omitempty"`
}

// GetIndex -
func (m Metadata) GetIndex() string {
	return models.DocTZIP
}

// GetScores -
func (m Metadata) GetScores(search string) []string {
	return []string{
		"name^15",
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
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return Item{
		Type:       m.GetIndex(),
		Value:      m.Address,
		Body:       m,
		Highlights: highlight,
		Network:    m.Network,
	}, nil
}
