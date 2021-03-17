package search

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
)

// Domain -
type Domain struct {
	Name       string            `json:"name"`
	Expiration time.Time         `json:"expiration"`
	Network    string            `json:"network"`
	Address    string            `json:"address"`
	Level      int64             `json:"level"`
	Data       map[string]string `json:"data,omitempty"`
}

// GetIndex -
func (d Domain) GetIndex() string {
	return models.DocTezosDomains
}

// GetScores -
func (d Domain) GetScores(search string) []string {
	return []string{
		"name^10",
		"address^5",
	}
}

// GetFields -
func (d Domain) GetFields() []string {
	return []string{
		"address",
		"name",
	}
}

// Parse  -
func (d Domain) Parse(highlight map[string][]string, data []byte) (interface{}, error) {
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return Item{
		Type:       d.GetIndex(),
		Value:      d.Address,
		Body:       d,
		Highlights: highlight,
		Network:    d.Network,
	}, nil
}
