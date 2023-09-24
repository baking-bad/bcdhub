package bigmapaction

import (
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

func (BigMapAction) PartitionBy() string {
	return ""
}
