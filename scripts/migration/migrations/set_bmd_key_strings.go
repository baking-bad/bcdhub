package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/metrics"
)

// SetBMDKeyStrings - migration that set key strings array at big map diff
type SetBMDKeyStrings struct{}

// Do - migrate function
func (m *SetBMDKeyStrings) Do(ctx *Context) error {
	log.Print("Start SetBMDKeyStrings migration...")
	start := time.Now()
	opIDs, err := ctx.ES.GetOperationsWithBigMapDiffs()
	if err != nil {
		return err
	}
	log.Printf("Found %d unique operations with big map diff", len(opIDs))

	h := metrics.New(ctx.ES, ctx.DB)
	for i := range opIDs {
		log.Printf("Compute for operation with id: %s", opIDs[i])
		if err := h.SetBigMapDiffsKeyString(opIDs[i]); err != nil {
			return err
		}
	}

	log.Printf("Time spent: %v", time.Since(start))

	return nil
}
