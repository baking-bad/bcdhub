package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/pkg/errors"
)

// CreateIndexers -
func CreateIndexers(ctx context.Context, cfg config.Config) ([]Indexer, error) {
	if err := tezerrors.LoadErrorDescriptions(); err != nil {
		return nil, err
	}

	indexers := make([]Indexer, 0)
	for network, indexerCfg := range cfg.Indexer.Networks {
		networkType := types.NewNetwork(network)
		if networkType == types.Empty {
			return nil, errors.Errorf("unknown network %s", network)
		}

		bi, err := NewBlockchainIndexer(ctx, cfg, networkType, indexerCfg)
		if err != nil {
			return nil, err
		}
		indexers = append(indexers, bi)
	}
	return indexers, nil
}
