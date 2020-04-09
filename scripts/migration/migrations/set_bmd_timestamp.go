package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/elastic"
)

// SetBMDTimestamp - migration that set timestamp at big map diff
type SetBMDTimestamp struct{}

// Do - migrate function
func (m *SetBMDTimestamp) Do(ctx *Context) error {
	log.Print("Start SetBMDTimestamp migration...")
	start := time.Now()
	allBMD, err := ctx.ES.GetAllBigMapDiff()
	if err != nil {
		return err
	}
	log.Printf("Found %d unique operations with big map diff", len(allBMD))

	ops := make(map[string]time.Time)
	var lastIdx int
	for i := range allBMD {
		log.Printf("Compute for bmd with id: %s", allBMD[i].ID)
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
			log.Printf("Saving updated data from %d to %d...", lastIdx, i)
			if err := ctx.ES.BulkUpdateBigMapDiffs(allBMD[lastIdx:i]); err != nil {
				return err
			}
			lastIdx = i
		}
	}

	log.Printf("Time spent: %v", time.Since(start))

	return nil
}
