package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/pkg/errors"
)

// CreateIndexers -
func CreateIndexers(ctx context.Context, internalCtx *config.Context, cfg config.Config) ([]Indexer, error) {
	if err := tezerrors.LoadErrorDescriptions(); err != nil {
		return nil, err
	}

	if err := NewInitializer(internalCtx.Storage, internalCtx.StorageDB.DB, cfg.Indexer.OffchainBaseURL).Init(ctx); err != nil {
		return nil, err
	}

	indexers := make([]Indexer, 0)
	for network := range cfg.Indexer.Networks {
		rpc, ok := cfg.RPC[network]
		if !ok {
			return nil, errors.Errorf("Unknown network %s", network)
		}

		bi, err := NewBoostIndexer(ctx, *internalCtx, rpc, types.NewNetwork(network))
		if err != nil {
			return nil, err
		}
		indexers = append(indexers, bi)
	}
	return indexers, nil
}
