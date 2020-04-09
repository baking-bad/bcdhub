package migrations

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
)

// SetBMDTimestamp - migration that set timestamp at big map diff
type SetBMDTimestamp struct{}

// Description -
func (m *SetBMDTimestamp) Description() string {
	return "set timestamp at big map diff"
}

// Do - migrate function
func (m *SetBMDTimestamp) Do(ctx *Context) error {
	allBMD, err := ctx.ES.GetAllBigMapDiff()
	if err != nil {
		return err
	}
	logger.Info("Found %d unique operations with big map diff", len(allBMD))

	ops := make(map[string]time.Time)
	var lastIdx int
	for i := range allBMD {
		logger.Info("Compute for bmd with id: %s", allBMD[i].ID)
		ts, ok := ops[allBMD[i].OperationID]
		if !ok {
			operation, err := ctx.ES.GetByID(elastic.DocOperations, allBMD[i].OperationID)
			if err != nil {
				return err
			}
			ts = operation.Get("_source.timestamp").Time().UTC()
		}
		allBMD[i].Timestamp = ts

		if (i%1000 == 0 || i == len(allBMD)-1) && i > 0 {
			logger.Info("Saving updated data from %d to %d...", lastIdx, i)
			if err := ctx.ES.BulkUpdateBigMapDiffs(allBMD[lastIdx:i]); err != nil {
				return err
			}
			lastIdx = i
		}
	}

	logger.Info("Done. Total operations: %d", len(allBMD))

	return nil
}
