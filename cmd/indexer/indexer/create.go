package indexer

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/config"
)

// CreateIndexers -
func CreateIndexers(ctx context.Context, cfg config.Config) ([]Indexer, error) {
	if err := tezerrors.LoadErrorDescriptions(); err != nil {
		return nil, err
	}

	indexers := make([]Indexer, 0)
	for network, indexerCfg := range cfg.Indexer.Networks {
		if indexerCfg.Periodic != nil {
			periodicIndexer, err := NewPeriodicIndexer(ctx, network, cfg, indexerCfg)
			if err != nil {
				return nil, err
			}
			indexers = append(indexers, periodicIndexer)
		} else {
			bi, err := NewBlockchainIndexer(ctx, cfg, network, indexerCfg)
			if err != nil {
				return nil, err
			}
			indexers = append(indexers, bi)
		}
	}
	return indexers, nil
}
