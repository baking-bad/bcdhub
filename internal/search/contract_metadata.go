package search

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	cm "github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
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

// NewMetadata -
func NewMetadata(network types.Network, contractMetadata *cm.ContractMetadata) Metadata {
	var m Metadata
	m.Address = contractMetadata.Address
	m.Authors = contractMetadata.Authors
	m.Description = contractMetadata.Description
	m.Homepage = contractMetadata.Homepage
	m.Level = contractMetadata.Level
	m.Name = contractMetadata.Name
	m.Network = network.String()
	m.Timestamp = contractMetadata.Timestamp.UTC()
	return m
}

// GetID -
func (m Metadata) GetID() string {
	return fmt.Sprintf("%s_%s", m.Network, m.Address)
}

// GetIndex -
func (m Metadata) GetIndex() string {
	return models.DocContractMetadata
}

// GetScores -
func (m Metadata) GetScores(search string) []string {
	return []string{
		"name^8",
		"authors^6",
		"homepage^6",
		"description^5",
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
func (m Metadata) Parse(highlight map[string][]string, data []byte) (*Item, error) {
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return &Item{
		Type:       m.GetIndex(),
		Value:      m.Address,
		Body:       &m,
		Highlights: highlight,
		Network:    m.Network,
	}, nil
}
