package indexer

import (
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/config"
)

// CreateIndexers -
func CreateIndexers(cfg config.Config) ([]Indexer, error) {
	if err := tezerrors.LoadErrorDescriptions(); err != nil {
		return nil, err
	}

	indexers := make([]Indexer, 0)
	for network, options := range cfg.Indexer.Networks {
		boostOptions := make([]BoostIndexerOption, 0)
		if options.Boost != "" {
			boostOptions = append(boostOptions, WithBoost(options.Boost, network, cfg))
		}
		if cfg.Indexer.SkipDelegatorBlocks {
			boostOptions = append(boostOptions, WithSkipDelegatorBlocks())
		}
		bi, err := NewBoostIndexer(cfg, network, boostOptions...)
		if err != nil {
			return nil, err
		}
		indexers = append(indexers, bi)
	}
	return indexers, nil
}
