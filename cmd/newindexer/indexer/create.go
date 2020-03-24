package indexer

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/index"
)

// CreateIndexers -
func CreateIndexers(cfg Config) ([]Indexer, error) {
	if err := cerrors.LoadErrorDescriptions("data/errors.json"); err != nil {
		return nil, err
	}
	return createIndexers(cfg)
}

func createIndexers(cfg Config) ([]Indexer, error) {
	indexers := make([]Indexer, 0)
	for network, config := range cfg.Indexers {
		if config.Boost {
			bi, err := NewBoostIndexer(cfg, network)
			if err != nil {
				return nil, err
			}
			indexers = append(indexers, bi)
		} else {
			// TODO: default indexer
			return nil, fmt.Errorf("Only `boost` indexer is supported now")
		}
	}
	return indexers, nil
}

func createExternalInexer(cfg *ExternalIndexerConfig) (index.Indexer, error) {
	if cfg == nil {
		return nil, fmt.Errorf("Empty `external_indexer` section in config. You have to set it when `boost` is true")
	}
	switch cfg.Type {
	case "tzkt":
		return index.NewTzKT(cfg.Link, time.Duration(cfg.Timeout)*time.Second), nil
	case "tzstats":
		return index.NewTzStats(cfg.Link), nil
	default:
		return nil, fmt.Errorf("Unknown external indexer type: %s", cfg.Type)
	}
}
