package migrations

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
)

// SetBMDTimestamp - migration that set timestamp at big map diff
type SetBMDTimestamp struct{}

// Description -
func (m *SetBMDTimestamp) Description() string {
	return "set timestamp at big map diff"
}

// Do - migrate function
func (m *SetBMDTimestamp) Do(ctx *config.Context) error {
	var allBMD []models.BigMapDiff
	if err := ctx.ES.GetAll(&allBMD); err != nil {
		return err
	}
	logger.Info("Found %d unique operations with big map diff", len(allBMD))

	ops := make(map[string]time.Time)
	var lastIdx int
	for i := range allBMD {
		logger.Info("Compute for bmd with id: %s", allBMD[i].ID)
		ts, ok := ops[allBMD[i].OperationID]
		if !ok {
			operation := models.Operation{ID: allBMD[i].OperationID}
			if err := ctx.ES.GetByID(&operation); err != nil {
				return err
			}
			ts = operation.Timestamp
		}
		allBMD[i].Timestamp = ts

		if (i%1000 == 0 || i == len(allBMD)-1) && i > 0 {
			logger.Info("Saving updated data from %d to %d...", lastIdx, i)
			updates := make([]elastic.Model, len(allBMD[lastIdx:i]))
			for j := range allBMD[lastIdx:i] {
				updates[j] = &allBMD[lastIdx:i][j]
			}
			if err := ctx.ES.BulkUpdate(updates); err != nil {
				return err
			}
			lastIdx = i
		}
	}

	logger.Info("Done. Total operations: %d", len(allBMD))

	return nil
}
