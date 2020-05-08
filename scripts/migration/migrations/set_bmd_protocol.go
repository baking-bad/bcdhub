package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/schollz/progressbar/v3"
)

// SetBMDProtocol - migration that set `Protocol` at big map diff
type SetBMDProtocol struct{}

// Description -
func (m *SetBMDProtocol) Description() string {
	return "set `Protocol` at big map diff"
}

// Do - migrate function
func (m *SetBMDProtocol) Do(ctx *config.Context) error {
	var allBMD []models.BigMapDiff
	if err := ctx.ES.GetAll(&allBMD); err != nil {
		return err
	}
	logger.Info("Found %d unique operations with big map diff", len(allBMD))

	bar := progressbar.NewOptions(len(allBMD), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())
	ops := make(map[string]string)
	var lastIdx int

	for i := range allBMD {
		bar.Add(1)

		proto, ok := ops[allBMD[i].OperationID]
		if !ok {
			operation := models.Operation{ID: allBMD[i].OperationID}
			if err := ctx.ES.GetByID(&operation); err != nil {
				return err
			}
			proto = operation.Protocol
		}
		allBMD[i].Protocol = proto

		if (i%1000 == 0 || i == len(allBMD)-1) && i > 0 {
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

	logger.Info("Done.")

	return nil
}
