package indexer

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/index"
)

// CreateIndexers -
func CreateIndexers(cfg Config) ([]Indexer, error) {
	if err := cerrors.LoadErrorDescriptions("data/errors.json"); err != nil {
		return nil, err
	}
	networks := make([]string, 0)
	for k := range cfg.Indexers {
		networks = append(networks, k)
	}
	es := elastic.WaitNew([]string{cfg.Search.URI})
	if err := meta.LoadProtocols(es, networks); err != nil {
		return nil, err
	}
	return createIndexers(cfg)
}

func createIndexers(cfg Config) ([]Indexer, error) {
	indexers := make([]Indexer, 0)
	for network := range cfg.Indexers {
		bi, err := NewBoostIndexer(cfg, network)
		if err != nil {
			return nil, err
		}
		indexers = append(indexers, bi)
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
	default:
		return nil, fmt.Errorf("Unknown external indexer type: %s", cfg.Type)
	}
}
