package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// SetOperationInitiator - migration that fill `Initiator` in operations
type SetOperationInitiator struct{}

// Key -
func (m *SetOperationInitiator) Key() string {
	return "set_operation_initiator"
}

// Description -
func (m *SetOperationInitiator) Description() string {
	return "fill `Initiator` in operations"
}

// Do - migrate function
func (m *SetOperationInitiator) Do(ctx *config.Context) error {
	start := time.Now()
	for _, network := range ctx.Config.Migrations.Networks {
		logger.Info("Starting %s...", network)
		updates := make([]elastic.Model, 0)

		operations, err := ctx.ES.GetOperations(
			map[string]interface{}{
				"network":  network,
				"internal": false,
			},
			0,
			false,
		)
		if err != nil {
			return err
		}
		logger.Info("Found %d main operations for %s", len(operations), network)

		hashMap := make(map[string]string)

		bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())
		for i := range operations {
			bar.Add(1) //nolint

			operations[i].Initiator = operations[i].Source
			hashMap[operations[i].Hash] = operations[i].Initiator
			updates = append(updates, &operations[i])
		}

		internalOperations, err := ctx.ES.GetOperations(
			map[string]interface{}{
				"network":  network,
				"internal": true,
			},
			0,
			false,
		)
		if err != nil {
			return err
		}
		logger.Info("Found %d internal operations for %s", len(internalOperations), network)

		bar = progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())
		for i := range internalOperations {
			bar.Add(1) //nolint

			if initiator, ok := hashMap[internalOperations[i].Hash]; ok {
				internalOperations[i].Initiator = initiator
				updates = append(updates, &internalOperations[i])
			}
		}
		logger.Info("Saving %d updated operations for %s", len(updates), network)

		if err := ctx.ES.BulkUpdate(updates); err != nil {
			return err
		}
	}

	log.Printf("Time spent: %v", time.Since(start))

	return nil
}
