package search

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
)

// Operation -
type Operation struct {
	Network  string `json:"network"`
	Protocol string `json:"protocol"`
	Hash     string `json:"hash"`
	Internal bool   `json:"internal"`

	Status           string    `json:"status"`
	Timestamp        time.Time `json:"timestamp"`
	Level            int64     `json:"level"`
	Kind             string    `json:"kind"`
	Initiator        string    `json:"initiator"`
	Source           string    `json:"source"`
	Destination      string    `json:"destination,omitempty"`
	FoundBy          string    `json:"found_by,omitempty"`
	Entrypoint       string    `json:"entrypoint,omitempty"`
	SourceAlias      string    `json:"source_alias,omitempty"`
	DestinationAlias string    `json:"destination_alias,omitempty"`

	DelegateAlias string `json:"delegate_alias,omitempty"`

	ParameterStrings []string `json:"parameter_strings,omitempty"`
	StorageStrings   []string `json:"storage_strings,omitempty"`
	Tags             []string `json:"tags,omitempty"`
}

// GetIndex -
func (o Operation) GetIndex() string {
	return models.DocOperations
}

// GetScores -
func (o Operation) GetScores(search string) []string {
	return []string{
		"entrypoint^8",
		"parameter_strings^7",
		"storage_strings^7",
		"errors.with^6",
		"errors.id^5",
		"source_alias^3",
		"hash",
		"source",
	}
}

// GetFields -
func (o Operation) GetFields() []string {
	return []string{
		"entrypoint",
		"parameter_strings",
		"storage_strings",
		"errors.with",
		"errors.id",
		"source_alias",
		"hash",
		"source",
	}
}

// Parse  -
func (o Operation) Parse(highlight map[string][]string, data []byte) (interface{}, error) {
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}
	return Item{
		Type:       o.GetIndex(),
		Value:      o.Hash,
		Body:       o,
		Highlights: highlight,
		Network:    o.Network,
	}, nil
}
