package migration

import (
	"github.com/baking-bad/bcdhub/internal/models"
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
func (storage *Storage) Get(network, address string) ([]migration.Migration, error) {
	var migrations []migration.Migration
	err := storage.DB.Table(models.DocMigrations).
		Scopes(
			core.NetworkAndAddress(network, address),
			core.OrderByLevelDesc,
		).
		Find(&migrations).Error
	return migrations, err
}

// Count -
func (storage *Storage) Count(network, address string) (int64, error) {
	var count int64
	err := storage.DB.Table(models.DocMigrations).
		Where("network = ?", network).
		Where("address = ?", address).
		Count(&count).
		Error
	return count, err
}
