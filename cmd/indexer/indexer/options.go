package indexer

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/index"
)

// BoostIndexerOption -
type BoostIndexerOption func(*BoostIndexer)

// WithBoost -
func WithBoost(externalType, network string, cfg config.Config) BoostIndexerOption {
	return func(bi *BoostIndexer) {
		if externalType == "" {
			return
		}

		bi.boost = true
		switch externalType {
		case "tzkt":
			bi.externalIndexer = index.NewTzKT(cfg.TzKT[network].URI, time.Duration(cfg.TzKT[network].Timeout)*time.Second)
			return
		default:
			panic(fmt.Errorf("unsupported external indexer type: %s", externalType))
		}
	}
}

// WithSkipDelegatorBlocks -
func WithSkipDelegatorBlocks() BoostIndexerOption {
	return func(bi *BoostIndexer) {
		bi.skipDelegatorBlocks = true
	}
}
