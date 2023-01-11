package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
)

// Initializer -
type Initializer struct {
	repo    models.GeneralRepository
	db      pg.DBI
	network types.Network
}

// NewInitializer -
func NewInitializer(network types.Network, repo models.GeneralRepository, db pg.DBI) Initializer {
	return Initializer{repo, db, network}
}

// Init -
func (initializer Initializer) Init(ctx context.Context) error {
	if err := initializer.repo.CreateTables(); err != nil {
		return err
	}

	return createStartIndices(initializer.db)
}
