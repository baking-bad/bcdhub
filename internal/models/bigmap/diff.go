package bigmap

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Diff -
type Diff struct {
	ID          int64       `json:"-"`
	Key         types.Bytes `json:"key" gorm:"type:bytes;not null"`
	KeyHash     string      `json:"key_hash" gorm:"index:big_map_diffs_key_hash_idx"`
	Value       types.Bytes `json:"value,omitempty" gorm:"type:bytes"`
	Level       int64       `json:"level"`
	Timestamp   time.Time   `json:"timestamp"`
	ProtocolID  int64       `json:"protocol" gorm:"type:SMALLINT"`
	OperationID int64       `json:"-" gorm:"index:big_map_diffs_operation_id_idx"`

	KeyStrings   pq.StringArray `json:"key_strings,omitempty" gorm:"type:text[]"`
	ValueStrings pq.StringArray `json:"value_strings,omitempty" gorm:"type:text[]"`

	BigMapID int64 `gorm:"not null;index:bm_diff_id_key_idx"`
	BigMap   BigMap
}

// TableName -
func (Diff) TableName() string {
	return "big_map_diffs"
}

// GetID -
func (b *Diff) GetID() int64 {
	return b.ID
}

// GetIndex -
func (b *Diff) GetIndex() string {
	return "big_map_diffs"
}

// Save -
func (b *Diff) Save(tx *gorm.DB) error {
	return tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(b).Error
}

// LogFields -
func (b *Diff) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network":  b.BigMap.Network.String(),
		"contract": b.BigMap.Contract,
		"ptr":      b.BigMap.Ptr,
		"block":    b.Level,
		"key_hash": b.KeyHash,
	}
}

// KeyBytes -
func (b *Diff) KeyBytes() []byte {
	if len(b.Key) >= 2 {
		if b.Key[0] == 34 && b.Key[len(b.Key)-1] == 34 {
			return b.Key[1 : len(b.Key)-1]
		}
	}
	return b.Key
}

// ValueBytes -
func (b *Diff) ValueBytes() []byte {
	if len(b.Value) >= 2 {
		if b.Value[0] == 34 && b.Value[len(b.Value)-1] == 34 {
			return b.Value[1 : len(b.Value)-1]
		}
	}
	return b.Value
}

// ToState -
func (b *Diff) ToState() *State {
	state := &State{
		BigMapID:        b.BigMapID,
		BigMap:          b.BigMap,
		LastUpdateLevel: b.Level,
		LastUpdateTime:  b.Timestamp,
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
