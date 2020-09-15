package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/parsers"
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
	h := metrics.New(ctx.ES, ctx.DB)
	operations, err := ctx.ES.GetOperations(map[string]interface{}{
		"entrypoint": "transfer",
		// "destination": "KT1KzoKR7v1HjF2JqfYAWFV2ihzmUVJsDqXy",
		// "network":     "mainnet",
	}, 0, false)
	if err != nil {
		return err
	}
	logger.Info("Found %d operations with transfer entrypoint", len(operations))

	tokenViews, err := parsers.NewTokenViews(ctx.DB)
	if err != nil {
		return err
	}

	result := make([]elastic.Model, 0)

	bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for i := range operations {
		if err := bar.Add(1); err != nil {
			return err
		}
		rpc, err := ctx.GetRPC(operations[i].Network)
		if err != nil {
			return err
		}

		parser := parsers.NewTransferParser(rpc, ctx.ES)
		parser.SetViews(tokenViews)

		transfers, err := parser.Parse(operations[i])
		if err != nil {
			return err
		}

		for i := range transfers {
			h.SetTransferAliases(ctx.Aliases, transfers[i])
			result = append(result, transfers[i])
		}
	}

	if err := ctx.ES.BulkInsert(result); err != nil {
		logger.Errorf("ctx.ES.BulkUpdate error: %v", err)
		return err
	}

	logger.Info("Done. %d transfers were saved.", len(result))

	return nil
}
