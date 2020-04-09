package migrations

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
)

// SetBMDKeyStrings - migration that set key strings array at big map diff
type SetBMDKeyStrings struct{}

// Description -
func (m *SetBMDKeyStrings) Description() string {
	return "set key strings array at big map diff"
}

// Do - migrate function
func (m *SetBMDKeyStrings) Do(ctx *Context) error {
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
		logger.Info("Compute for operation with id: %s", id)
		if err := h.SetBigMapDiffsKeyString(id); err != nil {
			return err
		}
	}

	logger.Info("Done. Total opunique operations: %d", len(opIDs))

	return nil
}
