package migration

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
)

// Storage -
type Storage struct {
	db *core.Reindexer
}

// NewStorage -
func NewStorage(db *core.Reindexer) *Storage {
	return &Storage{db}
}

// Get -
func (storage *Storage) Get(network, address string) (migrations []migration.Migration, err error) {
	query := storage.db.Query(models.DocMigrations).
		Match("network", network).
		Match("address", address).
		Sort("level", true)

	err = storage.db.GetAllByQuery(query, &migrations)
	return
}

// Count -
func (storage *Storage) Count(network, address string) (int64, error) {
	query := storage.db.Query(models.DocMigrations).
		Match("network", network).
		OpenBracket().
		Match("source", address).
		Or().
		Match("destination", address).
		CloseBracket()

	return storage.db.Count(query)
}
