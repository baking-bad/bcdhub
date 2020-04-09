package migrations

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/schollz/progressbar"
)

// SetTimestampMigration - migration that set timestamp from block head to operation
type SetTimestampMigration struct{}

// Do - migrate function
func (m *SetTimestampMigration) Do(ctx *Context) error {
	for _, network := range []string{consts.Mainnet, consts.Zeronet, consts.Carthage, consts.Babylon} {
		rpc, err := ctx.GetRPC(network)
		if err != nil {
			return err
		}

		operations, err := ctx.ES.GetAllOperations(network)
		if err != nil {
			return err
		}

		logger.Info("Found %d operations in %s", len(operations), network)

		bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false))

		lastLevel := int64(-1)
		var lastTs time.Time
		for _, operation := range operations {
			bar.Add(1)
			if operation.Level == lastLevel {
				operation.Timestamp = lastTs
			} else {
				ts, err := rpc.GetLevelTime(int(operation.Level))
				if err != nil {
					fmt.Print("\033[2K\r")
					return err
				}
				operation.Timestamp = ts
			}
			if _, err := ctx.ES.UpdateDoc(elastic.DocOperations, operation.ID, operation); err != nil {
				fmt.Print("\033[2K\r")
				return err
			}
			lastTs = operation.Timestamp
			lastLevel = operation.Level
		}

		fmt.Print("\033[2K\r")
		logger.Info("[%s] done. Total operations: %d", network, len(operations))
	}
	return nil
}
