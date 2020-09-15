package indexer

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
)

// CreateIndexers -
func CreateIndexers(cfg config.Config) ([]Indexer, error) {
	if err := cerrors.LoadErrorDescriptions("data/errors.json"); err != nil {
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
