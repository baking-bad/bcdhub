package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/dapp"
	"github.com/baking-bad/bcdhub/internal/parsers/contract_metadata"
	"github.com/go-pg/pg/v10"
)

// Initializer -
type Initializer struct {
	repo            models.GeneralRepository
	db              pg.DBI
	offchainBaseURL string
}

// NewInitializer -
func NewInitializer(repo models.GeneralRepository, db pg.DBI, offchainBaseURL string) Initializer {
	return Initializer{repo, db, offchainBaseURL}
}

// Init -
func (initializer Initializer) Init(ctx context.Context) error {
	if err := initializer.repo.CreateTables(); err != nil {
		return err
	}

	if initializer.offchainBaseURL != "" {
		count, err := initializer.db.Model((*dapp.DApp)(nil)).Count()
		if err != nil {
			return err
		}
		if count == 0 {
			offchainParser := contract_metadata.NewOffchain(initializer.offchainBaseURL)
			dapps, err := offchainParser.GetDApps(ctx)
			if err != nil {
				return err
			}
			if _, err := initializer.db.Model(&dapps).Returning("id").Insert(); err != nil {
				return err
			}

			metadata, err := offchainParser.GetContractMetadata(ctx)
			if err != nil {
				return err
			}
			if _, err := initializer.db.Model(&metadata).Returning("id").Insert(); err != nil {
				return err
			}
		}
	}

	return createStartIndices(initializer.db)
}
