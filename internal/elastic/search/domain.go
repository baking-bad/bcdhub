package search

import (
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
)

// Domain -
type Domain struct{}

// GetIndex -
func (d Domain) GetIndex() string {
	return "tezos_domain"
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
	var domain tezosdomain.TezosDomain
	if err := json.Unmarshal(data, &domain); err != nil {
		return nil, err
	}
	return Item{
		Type:       d.GetIndex(),
		Value:      domain.Address,
		Body:       domain,
		Highlights: highlight,
		Network:    domain.Network,
	}, nil
}
