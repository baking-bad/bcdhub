package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/schollz/progressbar/v3"
)

// CreateTransfersTags -
type CreateTransfersTags struct{}

// Key -
func (m *CreateTransfersTags) Key() string {
	return "create_transfers"
}

// Description -
func (m *CreateTransfersTags) Description() string {
	return "creates 'transfer' index"
}

// Do - migrate function
func (m *CreateTransfersTags) Do(ctx *config.Context) error {
	operations, err := ctx.ES.GetOperations(map[string]interface{}{
		"entrypoint": "transfer",
	}, 0, false)
	if err != nil {
		return err
	}
	logger.Info("Found %d operations with transfer entrypoint", len(operations))

	result := make([]elastic.Model, 0)

	bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for i := range operations {
		bar.Add(1)
		transfers, err := models.CreateTransfers(&operations[i])
		if err != nil {
			return err
		}
		for i := range transfers {
			result = append(result, transfers[i])
		}
	}

	if err := ctx.ES.BulkInsert(result); err != nil {
		logger.Errorf("ctx.ES.BulkUpdate error: %v", err)
		return err
	}

	logger.Info("Done. %d transfers were saves.", len(result))

	return nil
}
