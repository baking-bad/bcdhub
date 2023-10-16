package bigmapaction

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// Get -
func (storage *Storage) Get(ctx context.Context, ptr, limit, offset int64) (actions []bigmapaction.BigMapAction, err error) {
	query := storage.DB.NewSelect().Model(&actions).
		Where("source_ptr = ? AND action <> 3", ptr).WhereOr("destination_ptr = ? AND action = 3 ", ptr).
		Order("id DESC")

	if limit > 0 {
		query.Limit(int(limit))
	}
	if offset > 0 {
		query.Offset(int(offset))
	}
	err = query.Scan(ctx)
	return
}
