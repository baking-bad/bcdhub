package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/internal/logger"
	"gorm.io/gorm"
)

func timeBasedTask(period time.Duration, handler func() error, closeChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	for {
		select {
		case <-closeChan:
			return
		case <-ticker.C:
			if err := handler(); err != nil {
				logger.Err(err)
			}
		}
	}
}

func (ctx *Context) updateMaterializedViews() error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(`REFRESH MATERIALIZED VIEW CONCURRENTLY head_stats;`).Error
	})
}

func (ctx *Context) updateSeriesMaterializedViews() error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		for network := range ctx.Config.Indexer.Networks {

			if err := tx.Exec(fmt.Sprintf(`REFRESH MATERIALIZED VIEW CONCURRENTLY series_contract_by_month_%s;`, network)).Error; err != nil {
				return err
			}

			if err := tx.Exec(fmt.Sprintf(`REFRESH MATERIALIZED VIEW CONCURRENTLY series_operation_by_month_%s;`, network)).Error; err != nil {
				return err
			}

			if err := tx.Exec(fmt.Sprintf(`REFRESH MATERIALIZED VIEW CONCURRENTLY series_paid_storage_size_diff_by_month_%s;`, network)).Error; err != nil {
				return err
			}

			if err := tx.Exec(fmt.Sprintf(`REFRESH MATERIALIZED VIEW CONCURRENTLY series_consumed_gas_by_month_%s;`, network)).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
