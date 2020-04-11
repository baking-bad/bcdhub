package migrations

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
)

// SetBMDStrings - migration that set key and value strings array at big map diff
type SetBMDStrings struct{}

// Description -
func (m *SetBMDKeyStrings) Description() string {
	return "set key strings array at big map diff"
}

// Do - migrate function
<<<<<<< HEAD:scripts/migration/migrations/set_bmd_strings.go
func (m *SetBMDStrings) Do(ctx *Context) error {
	log.Print("Start SetBMDStrings migration...")
	start := time.Now()
=======
func (m *SetBMDKeyStrings) Do(ctx *Context) error {
>>>>>>> master:scripts/migration/migrations/set_bmd_key_strings.go
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
<<<<<<< HEAD:scripts/migration/migrations/set_bmd_strings.go
		log.Printf("Compute for operation with id: %s", id)
		if err := h.SetBigMapDiffsStrings(id); err != nil {
=======
		logger.Info("Compute for operation with id: %s", id)
		if err := h.SetBigMapDiffsKeyString(id); err != nil {
>>>>>>> master:scripts/migration/migrations/set_bmd_key_strings.go
			return err
		}
	}

	logger.Info("Done. Total opunique operations: %d", len(opIDs))

	return nil
}
