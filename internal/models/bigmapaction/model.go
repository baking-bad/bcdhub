package bigmapaction

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
)

// BigMapAction -
type BigMapAction struct {
	// nolint
	tableName struct{} `pg:"big_map_actions"`

	ID             int64
	Action         types.BigMapAction `pg:",type:SMALLINT"`
	SourcePtr      *int64
	DestinationPtr *int64
	OperationID    int64
	Level          int64
	Address        string
	Network        types.Network `pg:",type:SMALLINT"`
	Timestamp      time.Time
}

// GetID -
func (b *BigMapAction) GetID() int64 {
	return b.ID
}

// GetIndex -
func (b *BigMapAction) GetIndex() string {
	return "big_map_actions"
}

// Save -
func (b *BigMapAction) Save(tx pg.DBI) error {
	_, err := tx.Model(b).OnConflict("(id) DO UPDATE").
		Set(`
		action = excluded.action, 
		source_ptr = excluded.source_ptr, 
		destination_ptr = excluded.destination_ptr,
		operation_id = excluded.operation_id,
		level = excluded.level,
		address = excluded.address,
		network = excluded.network,
		timestamp = excluded.timestamp`).
		Returning("id").Insert()
	return err
}
