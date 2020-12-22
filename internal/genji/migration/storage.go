package migration

import (
	"github.com/baking-bad/bcdhub/internal/genji/core"
	"github.com/baking-bad/bcdhub/internal/models/migration"
)

// Storage -
type Storage struct {
	db *core.Genji
}

// NewStorage -
func NewStorage(db *core.Genji) *Storage {
	return &Storage{db}
}

// GetMigrations -
func (storage *Storage) GetMigrations(network, address string) (migrations []migration.Migration, err error) {
	builder := core.NewBuilder().And(
		core.NewEq("network", network),
		core.NewEq("address", address),
	).SortDesc("level")

	err = storage.db.GetAllByQuery(builder, &migrations)
	return
}
