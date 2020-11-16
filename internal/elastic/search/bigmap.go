package search

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/models"
)

// BigMap -
type BigMap struct{}

// GetIndex -
func (b BigMap) GetIndex() string {
	return "bigmapdiff"
}

// GetScores -
func (b BigMap) GetScores(search string) []string {
	return []string{
		"key_strings^8",
		"value_strings^7",
		"key_hash",
		"address",
	}
}

// GetFields -
func (b BigMap) GetFields() []string {
	return []string{
		"key_strings",
		"value_strings",
		"key_hash",
		"address",
	}
}

// Parse  -
func (b BigMap) Parse(highlight map[string][]string, data []byte) (interface{}, error) {
	var bmd models.BigMapDiff
	if err := json.Unmarshal(data, &bmd); err != nil {
		return nil, err
	}
	return Item{
		Type:       b.GetIndex(),
		Value:      bmd.KeyHash,
		Body:       bmd,
		Highlights: highlight,
		Network:    bmd.Network,
	}, nil
}
