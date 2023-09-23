package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/uptrace/bun"
)

// Initializer -
type Initializer struct {
	repo       models.GeneralRepository
	block      block.Repository
	db         bun.IDB
	network    types.Network
	rpc        noderpc.INode
	isPeriodic bool
}

// NewInitializer -
func NewInitializer(
	network types.Network,
	repo models.GeneralRepository,
	block block.Repository,
	db bun.IDB,
	rpc noderpc.INode,
	isPeriodic bool) Initializer {
	return Initializer{repo, block, db, network, rpc, isPeriodic}
}

// Init -
func (initializer Initializer) Init(ctx context.Context) error {
	if initializer.isPeriodic {
		if exists := initializer.repo.TablesExist(ctx); exists {
			// check first block in node and in database, compare its hash.
			// if hash is differed new periodic chain was started.
			logger.Info().Str("network", initializer.network.String()).Msg("checking for new periodic chain...")
			blockHash, err := initializer.rpc.BlockHash(ctx, 1)
			if err != nil {
				return err
			}
			firstBlock, err := initializer.block.Get(ctx, 1)
			if err == nil && firstBlock.Hash != blockHash {
				logger.Info().Str("network", initializer.network.String()).Msg("found new periodic chain")
				logger.Warning().Str("network", initializer.network.String()).Msg("drop database...")
				if err := initializer.repo.Drop(ctx); err != nil {
					return err
				}
				logger.Warning().Str("network", initializer.network.String()).Msg("database was dropped")
			}
		}
	}

	return initializer.repo.InitDatabase(ctx)
}
