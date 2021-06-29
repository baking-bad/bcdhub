package bigmap

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// State -
type State struct {
	ID              int64       `json:"-" gorm:"autoIncrement:true"`
	LastUpdateLevel int64       `json:"last_update_level" gorm:"last_update_level"`
	Count           int64       `json:"count" gorm:"default:0"`
	LastUpdateTime  time.Time   `json:"last_update_time"  gorm:"last_update_time"`
	KeyHash         string      `json:"key_hash" gorm:"not null;uniqueIndex:bm_state_key_idx"`
	Key             types.Bytes `json:"key" gorm:"type:bytes;not null"`
	Value           types.Bytes `json:"value,omitempty" gorm:"type:bytes"`
	Removed         bool        `json:"removed"`

	IsRollback bool `json:"-" gorm:"-"`

	BigMapID int64 `gorm:"not null;uniqueIndex:bm_state_key_idx"`
	BigMap   BigMap
}

// TableName -
func (State) TableName() string {
	return "big_map_states"
}

// GetID -
func (b *State) GetID() int64 {
	return b.ID
}

// GetIndex -
func (b *State) GetIndex() string {
	return "big_map_states"
}

// Save -
func (b *State) Save(tx *gorm.DB) error {
	var s clause.Set

	switch {
	case b.IsRollback:
		s = clause.Assignments(map[string]interface{}{
			"removed":           b.Removed,
			"value":             b.Value,
			"last_update_level": b.LastUpdateLevel,
			"last_update_time":  b.LastUpdateTime,
			"count":             gorm.Expr(`big_map_states."count"-1`),
		})
	case b.Removed:
		s = clause.Assignments(map[string]interface{}{
			"removed":           b.Removed,
			"last_update_level": b.LastUpdateLevel,
			"last_update_time":  b.LastUpdateTime,
			"count":             gorm.Expr(`big_map_states."count"+1`),
		})
	default:
		s = clause.Assignments(map[string]interface{}{
			"value":             b.Value,
			"last_update_level": b.LastUpdateLevel,
			"last_update_time":  b.LastUpdateTime,
			"count":             gorm.Expr(`big_map_states."count"+1`),
		})
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "big_map_id"},
			{Name: "key_hash"},
		},
		DoUpdates: s,
	}).Omit("BigMap").Create(b).Error
}

// LogFields -
func (b *State) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network":  b.BigMap.Network.String(),
		"ptr":      b.BigMap.Ptr,
		"key_hash": b.KeyHash,
		"removed":  b.Removed,
	}
}

// ToDiff -
func (b *State) ToDiff() Diff {
	bmd := Diff{
		BigMapID:  b.BigMapID,
		BigMap:    b.BigMap,
		KeyHash:   b.KeyHash,
		Key:       b.Key,
		Value:     b.Value,
		Level:     b.LastUpdateLevel,
		Timestamp: b.LastUpdateTime,
	}

	if b.Removed {
		bmd.Value = nil
	}

	return bmd
}
