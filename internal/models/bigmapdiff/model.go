package bigmapdiff

import (
	stdJSON "encoding/json"
	"time"

	"github.com/sirupsen/logrus"
)

// BigMapDiff -
type BigMapDiff struct {
	ID           string             `json:"-"`
	Ptr          int64              `json:"ptr"`
	Key          stdJSON.RawMessage `json:"key"`
	KeyHash      string             `json:"key_hash"`
	KeyStrings   []string           `json:"key_strings"`
	Value        stdJSON.RawMessage `json:"value,omitempty"`
	ValueStrings []string           `json:"value_strings"`
	OperationID  string             `json:"operation_id"`
	Level        int64              `json:"level"`
	Address      string             `json:"address"`
	Network      string             `json:"network"`
	IndexedTime  int64              `json:"indexed_time"`
	Timestamp    time.Time          `json:"timestamp"`
	Protocol     string             `json:"protocol"`

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

// KeyBytes -
func (b *BigMapDiff) KeyBytes() []byte {
	if len(b.Key) >= 2 {
		if b.Key[0] == 34 && b.Key[len(b.Key)-1] == 34 {
			return b.Key[1 : len(b.Key)-1]
		}
	}
	return b.Key
}

// ValueBytes -
func (b *BigMapDiff) ValueBytes() []byte {
	if len(b.Value) >= 2 {
		if b.Value[0] == 34 && b.Value[len(b.Value)-1] == 34 {
			return b.Value[1 : len(b.Value)-1]
		}
	}
	return b.Value
}
