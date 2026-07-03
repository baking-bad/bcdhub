package indexer

import (
	"context"
	"maps"
	"sync"

	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/dipdup-io/workerpool"
	"github.com/rs/zerolog/log"
)

// CreateIndexers -
func CreateIndexers(ctx context.Context, cfg config.Config, g workerpool.Group) ([]Indexer, error) {
	if err := tezerrors.LoadErrorDescriptions(); err != nil {
		return nil, err
	}

	var (
		mx       sync.Mutex
		indexers = make([]Indexer, 0)
		wg       = new(sync.WaitGroup)
	)

	for network, indexerCfg := range cfg.Indexer.Networks {
		networkCfg := cfg
		networkCfg.RPC = maps.Clone(cfg.RPC)

		wg.Add(1)
		go func(network string, indexerCfg config.IndexerConfig) {
			defer wg.Done()

			if indexerCfg.Periodic != nil {
				periodicIndexer, err := NewPeriodicIndexer(ctx, network, networkCfg, indexerCfg, g)
				if err != nil {
					log.Err(err).Msg("NewPeriodicIndexer")
					return
				}
				mx.Lock()
				indexers = append(indexers, periodicIndexer)
				mx.Unlock()

				g.GoCtx(ctx, periodicIndexer.Start)
			} else {
				bi, err := NewBlockchainIndexer(ctx, networkCfg, network, indexerCfg)
				if err != nil {
					log.Err(err).Msg("NewBlockchainIndexer")
					return
				}
				mx.Lock()
				indexers = append(indexers, bi)
				mx.Unlock()

				g.GoCtx(ctx, bi.Start)
			}
		}(network, indexerCfg)
	}

	wg.Wait()
	return indexers, nil
}
