package migrations

import (
	"log"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
)

// SetBMDStrings - migration that set key and value strings array at big map diff
type SetBMDStrings struct{}

// Description -
func (m *SetBMDStrings) Description() string {
	return "parse big map keys & values strings"
}

// Do - migrate function
func (m *SetBMDStrings) Do(ctx *config.Context) error {
	log.Print("Start SetBMDStrings migration...")
	allBigMapDiffs, err := ctx.ES.GetAllBigMapDiff()
	if err != nil {
		return err
	}
	logger.Info("Found %d big map diff", len(allBigMapDiffs))

	opIDs := make(map[string]struct{})
	for i := range allBigMapDiffs {
		opIDs[allBigMapDiffs[i].OperationID] = struct{}{}
	}

	logger.Info("Found %d unique operations with big map diff", len(opIDs))

	h := metrics.New(ctx.ES, ctx.DB)
	for id := range opIDs {
		log.Printf("Compute for operation with id: %s", id)
		if err := h.SetBigMapDiffsStrings(id); err != nil {
			return err
		}
	}

	logger.Info("Done. Total opunique operations: %d", len(opIDs))

	return nil
}
