package search

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
)

// BigMap -
type BigMap struct{}

// GetIndex -
func (b BigMap) GetIndex() string {
	return models.DocBigMapDiff
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
	var bmd bigmapdiff.BigMapDiff
	if err := json.Unmarshal(data, &bmd); err != nil {
		return nil, err
	}
	return models.Item{
		Type:       b.GetIndex(),
		Value:      bmd.KeyHash,
		Body:       bmd,
		Highlights: highlight,
		Network:    bmd.Network,
	}, nil
}
