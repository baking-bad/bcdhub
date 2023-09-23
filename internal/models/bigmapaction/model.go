package bigmapaction

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/uptrace/bun"
)

// BigMapAction -
type BigMapAction struct {
	bun.BaseModel `bun:"big_map_actions"`

	ID             int64              `bun:"id,pk,notnull,autoincrement"`
	Action         types.BigMapAction `bun:",type:SMALLINT"`
	SourcePtr      *int64
	DestinationPtr *int64
	OperationID    int64
	Level          int64
	Address        string
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
func (b *BigMapAction) Save(ctx context.Context, tx bun.IDB) error {
	_, err := tx.NewInsert().Model(b).On("CONFLICT (id) DO UPDATE").
		Set("action = excluded.action").
		Set("source_ptr = excluded.source_ptr").
		Set("destination_ptr = excluded.destination_ptr").
		Set("operation_id = excluded.operation_id").
		Set("level = excluded.level").
		Set("address = excluded.address").
		Set("timestamp = excluded.timestamp").
		Returning("id").
		Exec(ctx)
	return err
}

func (BigMapAction) PartitionBy() string {
	return ""
}
