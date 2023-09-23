package migration

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/migration"
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
func (storage *Storage) Get(ctx context.Context, contractID int64) (migrations []migration.Migration, err error) {
	err = storage.DB.
		NewSelect().
		Model(&migrations).
		Where("contract_id = ?", contractID).
		Order("id desc").
		Scan(ctx)
	return
}
