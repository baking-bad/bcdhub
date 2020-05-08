package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/schollz/progressbar/v3"
)

// SetOperationBurned - migration that set burned to operations in all networks
type SetOperationBurned struct{}

// Description -
func (m *SetOperationBurned) Description() string {
	return "set burned to operations in all networks"
}

// Do - migrate function
func (m *SetOperationBurned) Do(ctx *config.Context) error {
	h := metrics.New(ctx.ES, ctx.DB)

	for _, network := range ctx.Config.Migrations.Networks {
		var operations []models.Operation
		if err := ctx.ES.GetByNetwork(network, &operations); err != nil {
			return err
		}

		logger.Info("Found %d operations in %s", len(operations), network)

		bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())

		var changed int64
		var bulk []elastic.Model

		for i := range operations {
			bar.Add(1)

			h.SetOperationBurned(&operations[i])

			if operations[i].Burned > 0 {
				changed++
				bulk = append(bulk, &operations[i])
			}

			if len(bulk) == 1000 || (i == len(operations)-1 && len(bulk) > 0) {
				if err := ctx.ES.BulkUpdate(bulk); err != nil {
					return err
				}
				bulk = bulk[:0]
			}
		}

		logger.Info("[%s] done. Total operations: %d. Changed: %d", network, len(operations), changed)
	}
	return nil
}
