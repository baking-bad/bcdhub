package search

import (
	stdJSON "encoding/json"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
)

// BigMapDiff -
type BigMapDiff struct {
	Ptr          int64              `json:"ptr"`
	Key          stdJSON.RawMessage `json:"key"`
	KeyHash      string             `json:"key_hash"`
	Value        stdJSON.RawMessage `json:"value,omitempty"`
	OperationID  string             `json:"operation_id"`
	Level        int64              `json:"level"`
	Address      string             `json:"address"`
	Network      string             `json:"network"`
	Timestamp    time.Time          `json:"timestamp"`
	Protocol     string             `json:"protocol"`
	KeyStrings   []string           `json:"key_strings"`
	ValueStrings []string           `json:"value_strings"`
	FoundBy      string             `json:"found_by,omitempty"`
}

// GetIndex -
func (b BigMapDiff) GetIndex() string {
	return models.DocBigMapDiff
}

// GetScores -
func (b BigMapDiff) GetScores(search string) []string {
	return []string{
		"key_strings^8",
		"value_strings^7",
		"key_hash",
		"address",
	}
}

// GetFields -
func (b BigMapDiff) GetFields() []string {
	return []string{
		"key_strings",
		"value_strings",
		"key_hash",
		"address",
	}
}

// Parse  -
func (b BigMapDiff) Parse(highlight map[string][]string, data []byte) (interface{}, error) {
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, err
	}
	return Item{
		Type:       b.GetIndex(),
		Value:      b.KeyHash,
		Body:       b,
		Highlights: highlight,
		Network:    b.Network,
	}, nil
}
