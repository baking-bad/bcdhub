package models

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/utils"
	"github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

// BigMapDiff -
type BigMapDiff struct {
	ID           string      `json:"-"`
	Ptr          int64       `json:"ptr"`
	BinPath      string      `json:"bin_path"`
	Key          interface{} `json:"key"`
	KeyHash      string      `json:"key_hash"`
	KeyStrings   []string    `json:"key_strings"`
	Value        string      `json:"value"`
	ValueStrings []string    `json:"value_strings"`
	OperationID  string      `json:"operation_id"`
	Level        int64       `json:"level"`
	Address      string      `json:"address"`
	Network      string      `json:"network"`
	IndexedTime  int64       `json:"indexed_time"`
	Timestamp    time.Time   `json:"timestamp"`
	Protocol     string      `json:"protocol"`

	FoundBy string `json:"found_by,omitempty"`
}

// GetID -
func (b *BigMapDiff) GetID() string {
	return b.ID
}

// GetIndex -
func (b *BigMapDiff) GetIndex() string {
	return "bigmapdiff"
}

// GetQueues -
func (b *BigMapDiff) GetQueues() []string {
	return []string{"bigmapdiffs"}
}

// MarshalToQueue -
func (b *BigMapDiff) MarshalToQueue() ([]byte, error) {
	return []byte(b.ID), nil
}

// LogFields -
func (b *BigMapDiff) LogFields() logrus.Fields {
	return logrus.Fields{
		"network":  b.Network,
		"contract": b.Address,
		"ptr":      b.Ptr,
		"block":    b.Level,
		"key_hash": b.KeyHash,
	}
}

// GetScores -
func (b *BigMapDiff) GetScores(search string) []string {
	return []string{
		"key_strings^8",
		"value_strings^7",
		"key_hash",
		"address",
	}
}

// FoundByName -
func (b *BigMapDiff) FoundByName(hit gjson.Result) string {
	keys := hit.Get("highlight").Map()
	categories := b.GetScores("")
	return utils.GetFoundBy(keys, categories)
}
