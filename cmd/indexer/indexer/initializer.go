package indexer

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/rs/zerolog/log"
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
			log.Info().Str("network", initializer.network.String()).Msg("checking for new periodic chain...")

			var (
				notRunning = true
			)

			for notRunning {
				header, err := initializer.rpc.GetHead(ctx)
				if err != nil {
					return err
				}
				notRunning = header.Level == 0
				log.Info().Bool("running", !notRunning).Str("network", initializer.network.String()).Msg("chain status")
				if notRunning {
					time.Sleep(time.Second * 10)
				}
			}

			header, err := initializer.rpc.GetHeader(ctx, 1)
			if err != nil {
				return err
			}

			firstBlock, err := initializer.block.Get(ctx, 1)
			if err != nil {
				return nil
			}

			log.Info().
				Str("network", initializer.network.String()).
				Str("node_hash", header.Hash).
				Str("indexer_hash", firstBlock.Hash).
				Msg("checking first block hash...")
			if firstBlock.Hash != header.Hash {
				log.Info().Str("network", initializer.network.String()).Msg("found new periodic chain")
				if err := initializer.drop(ctx); err != nil {
					return err
				}
			}
		}
	}

	return initializer.repo.InitDatabase(ctx)
}

func (initializer Initializer) drop(ctx context.Context) error {
	log.Warn().Str("network", initializer.network.String()).Msg("drop database...")
	if err := initializer.repo.Drop(ctx); err != nil {
		return err
	}
	log.Warn().Str("network", initializer.network.String()).Msg("database was dropped")
	return nil
}
