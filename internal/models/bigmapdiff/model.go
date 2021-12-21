package bigmapdiff

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
	"github.com/lib/pq"
)

// BigMapDiff -
type BigMapDiff struct {
	// nolint
	tableName struct{} `pg:"big_map_diffs"`

	ID          int64
	Ptr         int64       `pg:",use_zero"`
	Key         types.Bytes `pg:",notnull,type:bytea"`
	KeyHash     string
	Value       types.Bytes `pg:",type:bytea"`
	Level       int64
	Contract    string
	Network     types.Network `pg:",type:SMALLINT"`
	Timestamp   time.Time
	ProtocolID  int64 `pg:",type:SMALLINT"`
	OperationID int64

	KeyStrings   pq.StringArray `pg:",type:text[]"`
	ValueStrings pq.StringArray ` pg:",type:text[]"`
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
func (b *BigMapDiff) Save(tx pg.DBI) error {
	_, err := tx.Model(b).
		OnConflict("(id) DO UPDATE").
		Set(`
			ptr = excluded.ptr, 
			key = excluded.key, 
			key_hash = excluded.key_hash, 
			value = excluded.value, 
			level = excluded.level, 
			contract = excluded.contract,
			network = excluded.network, 
			timestamp = excluded.timestamp, 
			protocol_id = excluded.protocol_id, 
			operation_id = excluded.operation_id, 
			key_strings = excluded.key_strings, 
			value_strings = excluded.value_strings`).
		Returning("id").
		Insert()
	return err
}

// LogFields -
func (b *BigMapDiff) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network":  b.Network.String(),
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
