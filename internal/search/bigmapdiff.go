package search

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/parsers/storage"
)

// BigMapDiff -
type BigMapDiff struct {
	ID           string    `json:"-"`
	Ptr          int64     `json:"ptr"`
	Key          string    `json:"key"`
	KeyHash      string    `json:"key_hash"`
	Level        int64     `json:"level"`
	Address      string    `json:"address"`
	Network      string    `json:"network"`
	Timestamp    time.Time `json:"timestamp"`
	KeyStrings   []string  `json:"key_strings"`
	ValueStrings []string  `json:"value_strings"`
	FoundBy      string    `json:"found_by,omitempty"`
}

// NewBigMapDiff -
func NewBigMapDiff(network types.Network, model *bigmapdiff.BigMapDiff) BigMapDiff {
	var b BigMapDiff
	b.ID = helpers.GenerateID()
	b.Address = model.Contract
	b.KeyHash = model.KeyHash
	b.Level = model.Level
	b.Network = network.String()
	b.Ptr = model.Ptr
	b.Timestamp = model.Timestamp.UTC()

	var data ast.UntypedAST
	if err := json.Unmarshal(model.Key, &data); err == nil {
		if key, err := data.Stringify(); err == nil {
			b.Key = key
		}
	}

	if keyStrings, err := storage.GetStrings(model.KeyBytes()); err == nil {
		b.KeyStrings = keyStrings
	}

	if model.Value != nil {
		valStrings, err := storage.GetStrings(model.ValueBytes())
		if err != nil {
			logger.Error().Err(err).Msg("storage.GetStrings")
		} else {
			b.ValueStrings = valStrings
		}
	}
	return b
}

// GetID -
func (b BigMapDiff) GetID() string {
	return b.ID
}

// GetIndex -
func (b BigMapDiff) GetIndex() string {
	return models.DocBigMapDiff
}

// GetScores -
func (b BigMapDiff) GetScores(search string) []string {
	return []string{
		"key_hash^10",
		"key_strings^6",
	}
}

// GetFields -
func (b BigMapDiff) GetFields() []string {
	return []string{
		"key_strings",
		"key_hash",
	}
}

// Parse  -
func (b BigMapDiff) Parse(highlight map[string][]string, data []byte) (*Item, error) {
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return &Item{
		Type:       b.GetIndex(),
		Value:      fmt.Sprintf("%d", b.Ptr),
		Body:       &b,
		Highlights: highlight,
		Network:    b.Network,
	}, nil
}
