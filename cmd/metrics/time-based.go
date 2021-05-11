package main

import (
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
				logger.Error(err)
			}
		}
	}
}

func (ctx *Context) updateMaterializedViews() error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Exec(`REFRESH MATERIALIZED VIEW head_stats;`).Error
	})
}
