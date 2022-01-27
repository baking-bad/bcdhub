package indexer

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/go-pg/pg/v10"
)

// Initializer -
type Initializer struct {
	repo models.GeneralRepository
	db   pg.DBI
}

// NewInitializer -
func NewInitializer(repo models.GeneralRepository, db pg.DBI) Initializer {
	return Initializer{repo, db}
}

// Init -
func (initializer Initializer) Init() error {
	if err := initializer.repo.CreateTables(); err != nil {
		return err
	}

	return createStartIndices(initializer.db)
}
