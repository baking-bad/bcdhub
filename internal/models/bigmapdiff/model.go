package bigmapdiff

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BigMapDiff -
type BigMapDiff struct {
	ID               int64       `json:"-"`
	Ptr              int64       `json:"ptr" gorm:"index:bmd_idx"`
	Key              types.Bytes `json:"key" gorm:"type:bytes;not null"`
	KeyHash          string      `json:"key_hash"`
	Value            types.Bytes `json:"value,omitempty" gorm:"type:bytes"`
	Level            int64       `json:"level"`
	Contract         string      `json:"contract" gorm:"index:bmd_idx"`
	Network          string      `json:"network" gorm:"index:bmd_idx"`
	Timestamp        time.Time   `json:"timestamp"`
	Protocol         string      `json:"protocol"`
	OperationHash    string      `json:"op_hash" gorm:"index:big_map_diffs_operation_hash_idx"`
	OperationCounter int64       `json:"op_counter" gorm:"index:big_map_diffs_operation_hash_idx"`
	OperationNonce   *int64      `json:"op_nonce" gorm:"index:big_map_diffs_operation_hash_idx"`

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

// Save -
func (b *BigMapDiff) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(b).Error
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
		"contract": b.Contract,
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

// ToState -
func (b *BigMapDiff) ToState() *BigMapState {
	state := &BigMapState{
		Network:         b.Network,
		Contract:        b.Contract,
		Ptr:             b.Ptr,
		LastUpdateLevel: b.Level,
		KeyHash:         b.KeyHash,
		Key:             b.KeyBytes(),
	}

	val := b.ValueBytes()
	if len(val) == 0 {
		state.Removed = true
	} else {
		state.Value = val
	}

	return state
}
