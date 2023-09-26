package stats

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/stats"
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
func (storage *Storage) Get(ctx context.Context) (stats stats.Stats, err error) {
	err = storage.DB.
		NewSelect().
		Model(&stats).
		Limit(1).
		Scan(ctx)
	return
}
