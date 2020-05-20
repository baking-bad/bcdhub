package migrations

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/tzkt"
)

// FindLostOperations -
type FindLostOperations struct {
	Network string
}

// Key -
func (m *FindLostOperations) Key() string {
	return "lost"
}

// Description -
func (m *FindLostOperations) Description() string {
	return "find lost blocks"
}

// Do - migrate function
func (m *FindLostOperations) Do(ctx *config.Context) error {
	for network, tzktProvider := range ctx.Config.TzKT {
		logger.Info("Start searching in %s...", network)

		api := tzkt.NewTzKT(tzktProvider.URI, time.Second*time.Duration(tzktProvider.Timeout))
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

		logger.Errorf("In %s skipped %d levels", network, len(skipped))
	}

	logger.Success("done")
	return nil
}
