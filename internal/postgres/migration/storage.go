package migration

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
func (storage *Storage) Get(network types.Network, address string) ([]migration.Migration, error) {
	var migrations []migration.Migration
	query := storage.DB.Model().Table(models.DocMigrations)
	core.NetworkAndAddress(network, address)(query)
	core.OrderByLevelDesc(query)
	err := query.Select(&migrations)
	return migrations, err
}
