package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/metrics"
)

// SetOperationStrings - migration that set storage and parameter strings array at operation
type SetOperationStrings struct{}

// Do - migrate function
func (m *SetOperationStrings) Do(ctx *Context) error {
	log.Print("Start SetOperationStrings migration...")
	start := time.Now()
	operations, err := ctx.ES.GetAllOperations()
	if err != nil {
		return err
	}
	log.Printf("Found %d operations", len(operations))

	var lastIdx int
	h := metrics.New(ctx.ES, ctx.DB)
	for i := range operations {
		log.Printf("Compute for operation with id: %s", operations[i].ID)
		h.SetOperationStrings(&operations[i])

		if (i%1000 == 0 || i == len(operations)-1) && i > 0 {
			log.Printf("Saving updated data from %d to %d...", lastIdx, i)
			if err := ctx.ES.BulkUpdateOperations(operations[lastIdx:i]); err != nil {
				return err
			}
			lastIdx = i
		}
	}

	log.Printf("Time spent: %v", time.Since(start))

	return nil
}
