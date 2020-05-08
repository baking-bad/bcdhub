package migrations

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"

	"github.com/schollz/progressbar/v3"
)

// SetTimestamp - migration that set timestamp from block head to operation
type SetTimestamp struct{}

// Description -
func (m *SetTimestamp) Description() string {
	return "set timestamp from block head to operation"
}

// Do - migrate function
func (m *SetTimestamp) Do(ctx *config.Context) error {
	for _, network := range ctx.Config.Migrations.Networks {
		rpc, err := ctx.GetRPC(network)
		if err != nil {
			return err
		}

		var operations []models.Operation
		if err := ctx.ES.GetByNetwork(network, &operations); err != nil {
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
