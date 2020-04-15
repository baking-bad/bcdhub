package migrations

import (
	"time"

	"github.com/baking-bad/bcdhub/cmd/indexer/indexer"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/tzkt"
)

// FindLostOperations -
type FindLostOperations struct {
	Network string
}

// Description -
func (m *FindLostOperations) Description() string {
	return "find lost blocks and index it"
}

func getIndexerConfig(ctx *Context) indexer.Config {
	cfg := indexer.Config{
		Search:         ctx.Config.Search,
		Mq:             ctx.Config.Mq,
		FilesDirectory: ctx.Config.FilesDirectory,
	}

	entities := map[string]indexer.EntityConfig{}
	for _, network := range []string{consts.Mainnet, consts.Babylon, consts.Carthage, consts.Zeronet} {
		entities[network] = indexer.EntityConfig{
			Boost: false,
			RPC: indexer.RPCConfig{
				URLs:    ctx.Config.NodeRPC[network],
				Timeout: 20,
			},
		}
	}

	cfg.Indexers = entities

	return cfg
}

// Do - migrate function
func (m *FindLostOperations) Do(ctx *Context) error {
	for _, network := range []string{consts.Mainnet, consts.Babylon, consts.Carthage, consts.Zeronet} {
		logger.Info("Start searching in %s...", network)

		api := tzkt.NewTzKTForNetwork(network, time.Minute)
		tzktLevels, err := api.GetAllContractOperationBlocks()
		if err != nil {
			return err
		}

		logger.Info("TzKT %s found %d levels", network, len(tzktLevels))

		bcdLevels, err := ctx.ES.GetAllLevelsForNetwork(network)
		if err != nil {
			return err
		}
		logger.Info("BCD %s found %d levels", network, len(bcdLevels))

		skipped := make([]int64, 0)
		for _, tzktLevel := range tzktLevels {
			if _, ok := bcdLevels[tzktLevel]; !ok && tzktLevel != 1 {
				skipped = append(skipped, tzktLevel)
			}
		}

		logger.Warning("In %s skipped %d levels", network, len(skipped))

		configIndexer := getIndexerConfig(ctx)
		bi, err := indexer.NewBoostIndexer(configIndexer, network)
		if err != nil {
			return err
		}

		if err := bi.Index(skipped); err != nil {
			return err
		}
	}

	logger.Success("done")

	return nil
}
