package bigmapdiff

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
)

// BigMapState -
type BigMapState struct {
	// nolint
	tableName struct{} `pg:"big_map_states"`

	ID              int64         `pg:",pk"`
	Ptr             int64         `pg:",notnull,unique:big_map_key,use_zero"`
	LastUpdateLevel int64         `pg:"last_update_level"`
	Count           int64         `pg:",use_zero"`
	LastUpdateTime  time.Time     `pg:"last_update_time"`
	Network         types.Network `pg:",type:SMALLINT,notnull,unique:big_map_key,use_zero"`
	KeyHash         string        `pg:",notnull,unique:big_map_key"`
	Contract        string        `pg:",notnull,unique:big_map_key"` // contract is in primary key for supporting alpha protocol (mainnet before babylon)
	Key             types.Bytes   `pg:",type:bytea,notnull"`
	Value           types.Bytes   `pg:",type:bytea"`
	Removed         bool          `pg:",use_zero"`

	IsRollback bool `pg:"-"`
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
func (b *BigMapState) Save(tx pg.DBI) error {
	_, err := tx.
		Model(b).
		OnConflict("(network, contract, ptr, key_hash) DO UPDATE").
		Set("removed = EXCLUDED.removed, last_update_level = EXCLUDED.last_update_level, last_update_time = EXCLUDED.last_update_time, count = EXCLUDED.count + 1, value = CASE WHEN EXCLUDED.removed THEN big_map_state.value ELSE EXCLUDED.value END").
		Returning("id").
		Insert(b)
	return err
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
