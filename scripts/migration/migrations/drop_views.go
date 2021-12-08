package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"gorm.io/gorm"
)

// DropViews -
type DropViews struct{}

// Key -
func (m *DropViews) Key() string {
	return "drop_views"
}

// Description -
func (m *DropViews) Description() string {
	return "drop stats views"
}

// Do - migrate function
func (m *DropViews) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DROP MATERIALIZED VIEW IF EXISTS head_stats").Error; err != nil {
			return err
		} else {
			logger.Info().Msgf("head_stats was removed")
		}

		for network := range ctx.Config.Indexer.Networks {
			for _, view := range []string{
				"series_contract_by_month_",
				"series_operation_by_month_",
				"series_paid_storage_size_diff_by_month_",
				"series_consumed_gas_by_month_",
			} {
				name := fmt.Sprintf("%s%s", view, network)
				if err := tx.Exec("DROP MATERIALIZED VIEW IF EXISTS ?", gorm.Expr(name)).Error; err != nil {
					return err
				} else {
					logger.Info().Msgf("%s was removed", name)
				}
			}
		}

		return nil
	})
}
