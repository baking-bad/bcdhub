package search

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
)

// Contract -
type Contract struct {
	Network   string    `json:"network"`
	Level     int64     `json:"level"`
	Timestamp time.Time `json:"timestamp"`
	Language  string    `json:"language,omitempty"`
	Hash      string    `json:"hash"`

	Tags        []string `json:"tags,omitempty"`
	Hardcoded   []string `json:"hardcoded,omitempty"`
	FailStrings []string `json:"fail_strings,omitempty"`
	Annotations []string `json:"annotations,omitempty"`
	Entrypoints []string `json:"entrypoints,omitempty"`

	Address  string `json:"address"`
	Manager  string `json:"manager,omitempty"`
	Delegate string `json:"delegate,omitempty"`

	ProjectID     string `json:"project_id"`
	FoundBy       string `json:"found_by,omitempty"`
	Alias         string `json:"alias,omitempty"`
	DelegateAlias string `json:"delegate_alias,omitempty"`

	TxCount    *int64     `json:"tx_count,omitempty"`
	LastAction *time.Time `json:"last_action,omitempty"`
}

// GetID -
func (c *Contract) GetID() string {
	return fmt.Sprintf("%s_%s", c.Network, c.Address)
}

// GetIndex -
func (c Contract) GetIndex() string {
	return models.DocContracts
}

// GetScores -
func (c Contract) GetScores(search string) []string {
	if helpers.IsAddress(search) {
		return []string{
			"contract^10",
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
		"contract",
	}
}

// GetFields -
func (c Contract) GetFields() []string {
	return []string{
		"contract",
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
func (c Contract) Parse(highlight map[string][]string, data []byte) (*Item, error) {
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &Item{
		Type:       c.GetIndex(),
		Value:      c.Address,
		Body:       &c,
		Highlights: highlight,
		Network:    c.Network,
	}, nil
}

// Prepare -
func (c *Contract) Prepare(model models.Model) {
	cont, ok := model.(*contract.Contract)
	if !ok {
		return
	}

	c.Address = cont.Address
	c.Annotations = cont.Annotations
	c.Delegate = cont.Delegate
	c.Entrypoints = cont.Entrypoints
	c.FailStrings = cont.FailStrings
	c.Hardcoded = cont.Hardcoded
	c.Hash = cont.Hash
	c.Language = cont.Language
	c.Level = cont.Level
	c.Manager = cont.Manager
	c.Network = cont.Network.String()
	c.ProjectID = cont.ProjectID
	c.Tags = cont.Tags.ToArray()
	c.Timestamp = cont.Timestamp.UTC()
}
