package search

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
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

// GetID -
func (b *BigMapDiff) GetID() string {
	return b.ID
}

// GetIndex -
func (b *BigMapDiff) GetIndex() string {
	return models.DocBigMapDiff
}

// GetScores -
func (b BigMapDiff) GetScores(search string) []string {
	return []string{
		"key_hash^10",
		"key_strings^6",
		"value_strings^5",
	}
}

// GetFields -
func (b BigMapDiff) GetFields() []string {
	return []string{
		"key_strings",
		"value_strings",
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
		Value:      b.KeyHash,
		Body:       &b,
		Highlights: highlight,
		Network:    b.Network,
	}, nil
}

// Prepare -
func (b *BigMapDiff) Prepare(model models.Model) {
	bmd, ok := model.(*bigmapdiff.BigMapDiff)
	if !ok {
		return
	}

	b.ID = helpers.GenerateID()
	b.Address = bmd.Contract

	var data ast.UntypedAST
	if err := json.Unmarshal(bmd.Key, &data); err != nil {
		return
	}

	key, err := data.Stringify()
	if err != nil {
		return
	}

	b.Key = key
	b.KeyHash = bmd.KeyHash
	b.KeyStrings = bmd.KeyStrings
	b.Level = bmd.Level
	b.Network = bmd.Network.String()
	b.Ptr = bmd.Ptr
	b.Timestamp = bmd.Timestamp.UTC()
	b.ValueStrings = bmd.ValueStrings
}
