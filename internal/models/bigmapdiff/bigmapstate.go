package bigmapdiff

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/uptrace/bun"
)

// BigMapState -
type BigMapState struct {
	bun.BaseModel `bun:"big_map_states"`

	ID              int64       `bun:"id,pk,notnull,autoincrement"`
	Ptr             int64       `bun:"ptr,notnull,unique:big_map_state_unique"`
	LastUpdateLevel int64       `bun:"last_update_level"`
	Count           int64       `bun:"count"`
	LastUpdateTime  time.Time   `bun:"last_update_time"`
	KeyHash         string      `bun:"key_hash,type:text,notnull,unique:big_map_state_unique"`
	Contract        string      `bun:"contract,type:text,notnull,unique:big_map_state_unique"`
	Key             types.Bytes `bun:"key,type:bytea,notnull"`
	Value           types.Bytes `bun:"value,type:bytea"`
	Removed         bool        `bun:"removed"`

	IsRollback bool `bun:"-"`
}

// GetID -
func (b *BigMapState) GetID() int64 {
	return b.ID
}

func (BigMapState) TableName() string {
	return "big_map_states"
}

// LogFields -
func (b *BigMapState) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"ptr":      b.Ptr,
		"key_hash": b.KeyHash,
		"removed":  b.Removed,
	}
}

// ToDiff -
func (b *BigMapState) ToDiff() BigMapDiff {
	bmd := BigMapDiff{
		Ptr:       b.Ptr,
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
