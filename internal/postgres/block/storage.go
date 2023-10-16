package block

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/block"
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
func (storage *Storage) Get(ctx context.Context, level int64) (block block.Block, err error) {
	err = storage.DB.NewSelect().
		Model(&block).
		Where("level = ?", level).
		Limit(1).
		Relation("Protocol").
		Scan(ctx)
	return
}

// Last - returns current indexer state for network
func (storage *Storage) Last(ctx context.Context) (block block.Block, err error) {
	err = storage.DB.NewSelect().
		Model(&block).
		Order("id desc").
		Limit(1).
		Relation("Protocol").
		Scan(ctx)
	if storage.IsRecordNotFound(err) {
		err = nil
	}
	return
}
