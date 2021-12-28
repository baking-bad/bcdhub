package migration

import (
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
func (storage *Storage) Get(contractID int64) (migrations []migration.Migration, err error) {
	err = storage.DB.Model(&migrations).Where("contract_id = ?", contractID).Order("id desc").Select(&migrations)
	return
}
