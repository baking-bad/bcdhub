package indexer

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/index"
)

// CreateIndexers -
func CreateIndexers(cfg config.Config) ([]Indexer, error) {
	if err := cerrors.LoadErrorDescriptions("data/errors.json"); err != nil {
		return nil, err
	}
	indexers := make([]Indexer, 0)
	for network, options := range cfg.Indexer.Networks {
		bi, err := NewBoostIndexer(cfg, network, options.Boost)
		if err != nil {
			return nil, err
		}
		indexers = append(indexers, bi)
	}
	return indexers, nil
}

func createExternalIndexer(cfg config.Config, network, externalType string) (index.Indexer, error) {
	switch externalType {
	case "tzkt":
		return index.NewTzKT(cfg.TzKT[network].URI, time.Duration(cfg.TzKT[network].Timeout)*time.Second), nil
	default:
		return nil, fmt.Errorf("Unsupported external indexer type: %s", externalType)
	}
}
