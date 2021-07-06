package bigmapdiff

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BigMapState -
type BigMapState struct {
	ID              int64         `json:"-" gorm:"autoIncrement:true"`
	Ptr             int64         `json:"ptr" gorm:"not null;primaryKey;autoIncrement:false;index:big_map_state_ptr_idx"`
	LastUpdateLevel int64         `json:"last_update_level" gorm:"last_update_level"`
	Count           int64         `json:"count" gorm:"default:0"`
	LastUpdateTime  time.Time     `json:"last_update_time"  gorm:"last_update_time"`
	Network         types.Network `json:"network" gorm:"type:SMALLINT;not null;primaryKey;default:0;index:big_map_state_ptr_idx"`
	KeyHash         string        `json:"key_hash" gorm:"not null;primaryKey"`
	Contract        string        `json:"contract" gorm:"not null;primaryKey"` // contract is in primary key for supporting alpha protocol (mainnet before babylon)
	Key             types.Bytes   `json:"key" gorm:"type:bytes;not null"`
	Value           types.Bytes   `json:"value,omitempty" gorm:"type:bytes"`
	Removed         bool          `json:"removed"`

	IsRollback bool `json:"-" gorm:"-"`
}

// GetID -
func (b *BigMapState) GetID() int64 {
	return b.ID
}

// GetIndex -
func (b *BigMapState) GetIndex() string {
	return "big_map_states"
}

// Save -
func (b *BigMapState) Save(tx *gorm.DB) error {
	var s clause.Set

	switch {
	case b.IsRollback:
		s = clause.Assignments(map[string]interface{}{
			"removed":           b.Removed,
			"value":             b.Value,
			"last_update_level": b.LastUpdateLevel,
			"last_update_time":  b.LastUpdateTime,
			"count":             gorm.Expr(`big_map_states."count"+1`),
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
			{Name: "network"},
			{Name: "contract"},
			{Name: "ptr"},
			{Name: "key_hash"},
		},
		DoUpdates: s,
	}).Create(b).Error
}

// GetQueues -
func (b *BigMapState) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (b *BigMapState) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// LogFields -
func (b *BigMapState) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network":  b.Network.String(),
		"ptr":      b.Ptr,
		"key_hash": b.KeyHash,
		"removed":  b.Removed,
	}
}

// ToDiff -
func (b *BigMapState) ToDiff() BigMapDiff {
	bmd := BigMapDiff{
		Ptr:       b.Ptr,
		Network:   b.Network,
		KeyHash:   b.KeyHash,
		Contract:  b.Contract,
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
