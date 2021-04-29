package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"gorm.io/gorm"
)

// AddCountColumnToBigMapState -
type AddCountColumnToBigMapState struct{}

// Key -
func (m *AddCountColumnToBigMapState) Key() string {
	return "count_in_big_map_state"
}

// Description -
func (m *AddCountColumnToBigMapState) Description() string {
	return "add column `count` to `big_map_states` table"
}

// Do - migrate function
func (m *AddCountColumnToBigMapState) Do(ctx *config.Context) error {
	if err := ctx.StorageDB.DB.AutoMigrate(&bigmapdiff.BigMapState{}); err != nil {
		return err
	}

	limit := 100
	var offset int
	var end bool
	for !end {
		var states []bigmapdiff.BigMapState
		if err := ctx.StorageDB.DB.Table(models.DocBigMapState).Limit(limit).Offset(offset).Find(&states).Error; err != nil {
			return err
		}

		err := ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
			for i := range states {
				result := map[string]interface{}{}
				if err := ctx.StorageDB.DB.Table(models.DocBigMapDiff).
					Select(`count(*) as updates_count, max("timestamp") as last_update`).
					Where("network = ?", states[i].Network).
					Where("contract = ?", states[i].Contract).
					Where("key_hash = ?", states[i].KeyHash).
					Take(&result).Error; err != nil {
					return err
				}

				if err := tx.Model(&states[i]).Updates(map[string]interface{}{
					"count":            result["updates_count"],
					"last_update_time": result["last_update"],
				}).Error; err != nil {
					return err
				}
			}

			return nil
		})
		if err != nil {
			return err
		}

		offset += limit
		end = len(states) < limit

		logger.Info("Processed %d states", offset)
	}
	return nil
}
