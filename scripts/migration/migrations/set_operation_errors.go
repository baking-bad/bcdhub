package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// SetOperationErrors - migration that fill `Errors` in operations
type SetOperationErrors struct {
	Network string
}

// Key -
func (m *SetOperationErrors) Key() string {
	return "set_operation_errors"
}

// Description -
func (m *SetOperationErrors) Description() string {
	return "fill `Errors` in operations"
}

// Do - migrate function
func (m *SetOperationErrors) Do(ctx *config.Context) error {
	start := time.Now()
	for _, network := range ctx.Config.Migrations.Networks {
		operations, err := ctx.ES.GetOperations(
			map[string]interface{}{
				"network": network,
				"status":  "failed",
			},
			false,
			false,
		)
		if err != nil {
			return err
		}
		logger.Info("Found %d operations for %s", len(operations), network)

		var lastIdx int
		bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())
		for i := range operations {
			bar.Add(1) //nolint

			if (i%1000 == 0 || i == len(operations)-1) && i > 0 {
				updates := make([]elastic.Model, len(operations[lastIdx:i]))
				for j := range operations[lastIdx:i] {
					updates[j] = &operations[lastIdx:i][j]
				}
				if err := ctx.ES.BulkUpdate(updates); err != nil {
					return err
				}
				lastIdx = i
			}
		}
	}

	log.Printf("Time spent: %v", time.Since(start))

	return nil
}
