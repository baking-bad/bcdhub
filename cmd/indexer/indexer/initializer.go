package indexer

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/tzkt"
	"github.com/go-pg/pg/v10"
)

// Initializer -
type Initializer struct {
	repo    models.GeneralRepository
	db      pg.DBI
	tzktURI string
	network types.Network
}

// NewInitializer -
func NewInitializer(network types.Network, repo models.GeneralRepository, db pg.DBI, tzktURI string) Initializer {
	return Initializer{repo, db, tzktURI, network}
}

// Init -
func (initializer Initializer) Init(ctx context.Context) error {
	if err := initializer.repo.CreateTables(); err != nil {
		return err
	}

	return createStartIndices(initializer.db)
}

func (initializer *Initializer) getAliases(ctx context.Context) error {
	if initializer.tzktURI == "" {
		return nil
	}

	accounts, err := tzkt.NewTzKT(initializer.tzktURI, 10*time.Second).GetAliases()
	if err != nil {
		return err
	}

	return initializer.db.RunInTransaction(ctx, func(tx *pg.Tx) error {
		for address, alias := range accounts {
			acc := account.Account{
				Address: address,
				Type:    types.NewAccountType(address),
				Alias:   alias,
			}

			if _, err := initializer.db.Model(&acc).
				OnConflict("(address) DO NOTHING").
				Insert(); err != nil {
				return err
			}
		}

		return nil
	})
}
