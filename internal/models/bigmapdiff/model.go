package bigmapdiff

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

// BigMapDiff -
type BigMapDiff struct {
	ID               int64       `json:"-"`
	Ptr              int64       `json:"ptr"`
	Key              types.Bytes `json:"key" gorm:"type:bytes;not null"`
	KeyHash          string      `json:"key_hash"`
	Value            types.Bytes `json:"value,omitempty" gorm:"type:bytes"`
	Level            int64       `json:"level"`
	Address          string      `json:"address"`
	Network          string      `json:"network"`
	IndexedTime      int64       `json:"indexed_time"`
	Timestamp        time.Time   `json:"timestamp"`
	Protocol         string      `json:"protocol"`
	OperationHash    string      `json:"op_hash"`
	OperationCounter int64       `json:"op_counter"`
	OperationNonce   *int64      `json:"op_nonce"`

	KeyStrings   pq.StringArray `json:"key_strings,omitempty" gorm:"type:text[]"`
	ValueStrings pq.StringArray `json:"value_strings,omitempty" gorm:"type:text[]"`
}

// GetID -
func (b *BigMapDiff) GetID() int64 {
	return b.ID
}

// GetIndex -
func (b *BigMapDiff) GetIndex() string {
	return "big_map_diffs"
}

// GetQueues -
func (b *BigMapDiff) GetQueues() []string {
	return []string{"bigmapdiffs"}
}

// MarshalToQueue -
func (b *BigMapDiff) MarshalToQueue() ([]byte, error) {
	return []byte(fmt.Sprintf("%d", b.ID)), nil
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
