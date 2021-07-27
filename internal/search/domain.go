package search

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
)

// Domain -
type Domain struct {
	Name       string    `json:"name"`
	Expiration time.Time `json:"expiration"`
	Network    string    `json:"network"`
	Address    string    `json:"address"`
	Level      int64     `json:"level"`
	Timestamp  time.Time `json:"timestamp"`
}

// GetID -
func (d *Domain) GetID() string {
	return d.Name
}

// GetIndex -
func (d *Domain) GetIndex() string {
	return models.DocTezosDomains
}

// GetScores -
func (d Domain) GetScores(search string) []string {
	return []string{
		"name^8"
	}
}

// GetFields -
func (d Domain) GetFields() []string {
	return []string{
		"name",
	}
}

// Parse  -
func (d Domain) Parse(highlight map[string][]string, data []byte) (*Item, error) {
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, err
	}
	return &Item{
		Type:       d.GetIndex(),
		Value:      d.Address,
		Body:       &d,
		Highlights: highlight,
		Network:    d.Network,
	}, nil
}

// Prepare -
func (d *Domain) Prepare(model models.Model) {
	td, ok := model.(*tezosdomain.TezosDomain)
	if !ok {
		return
	}

	d.Address = td.Address
	d.Expiration = td.Expiration
	d.Level = td.Level
	d.Name = td.Name
	d.Network = td.Network.String()
	d.Timestamp = td.Timestamp.UTC()
}
