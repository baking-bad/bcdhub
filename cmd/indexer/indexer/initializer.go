package indexer

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	cmModel "github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/parsers/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/tzkt"
	"github.com/go-pg/pg/v10"
)

// Initializer -
type Initializer struct {
	repo            models.GeneralRepository
	db              pg.DBI
	tzktURI         string
	offchainBaseURL string
	network         types.Network
}

// NewInitializer -
func NewInitializer(network types.Network, repo models.GeneralRepository, db pg.DBI, offchainBaseURL, tzktURI string) Initializer {
	return Initializer{repo, db, tzktURI, offchainBaseURL, network}
}

// Init -
func (initializer Initializer) Init(ctx context.Context) error {
	if err := initializer.repo.CreateTables(); err != nil {
		return err
	}

	if initializer.offchainBaseURL != "" && initializer.network == types.Mainnet {
		count, err := initializer.db.Model((*cmModel.ContractMetadata)(nil)).Context(ctx).Count()
		if err != nil {
			return err
		}
		if count == 0 {
			logger.Info().Msg("loading offchain metadata...")
			offchainParser := contract_metadata.NewOffchain(initializer.offchainBaseURL)

			metadata, err := offchainParser.GetContractMetadata(ctx)
			if err != nil {
				return err
			}
			if _, err := initializer.db.Model(&metadata.Contracts).Context(ctx).Returning("id").Insert(); err != nil {
				return err
			}
			if _, err := initializer.db.Model(&metadata.Accounts).Context(ctx).Returning("id").Insert(); err != nil {
				return err
			}
			if _, err := initializer.db.Model(&metadata.Tokens).Context(ctx).Returning("id").Insert(); err != nil {
				return err
			}

			logger.Info().Msg("loading aliases...")
			if err := initializer.getAliases(ctx); err != nil {
				return nil
			}
		}
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
