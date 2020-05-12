package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/schollz/progressbar/v3"
)

// RecalcContractMetrics - migration that recalculate contract metrics
type RecalcContractMetrics struct{}

// Description -
func (m *RecalcContractMetrics) Description() string {
	return "recalculate contract metrics"
}

// Do - migrate function
func (m *RecalcContractMetrics) Do(ctx *config.Context) error {
	logger.Info("Start RecalcContractMetrics migration...")
	start := time.Now()
	h := metrics.New(ctx.ES, ctx.DB)

	for _, network := range ctx.Config.Migrations.Networks {
		contracts, err := ctx.ES.GetContracts(map[string]interface{}{
			"network": network,
		})
		if err != nil {
			return err
		}

		logger.Info("Found %d contracts in %s", len(contracts), network)

		bar := progressbar.NewOptions(len(contracts), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())

		var lastIdx int
		for i := range contracts {
			bar.Add(1)

			if err := h.UpdateContractStats(&contracts[i]); err != nil {
				return err
			}

			if (i%1000 == 0 || i == len(contracts)-1) && i > 0 {
				updates := make([]elastic.Model, len(contracts[lastIdx:i]))
				for j := range contracts[lastIdx:i] {
					updates[j] = &contracts[lastIdx:i][j]
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
