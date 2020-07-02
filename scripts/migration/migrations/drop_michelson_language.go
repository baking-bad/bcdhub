package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/language"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// DropMichelson - migration that drops michelson langugage and set it to unknown
type DropMichelson struct{}

// Key -
func (m *DropMichelson) Key() string {
	return "drop_michelson"
}

// Description -
func (m *DropMichelson) Description() string {
	return "drop michelson langugage and set it to unknown"
}

// Do - migrate function
func (m *DropMichelson) Do(ctx *config.Context) error {
	filter := make(map[string]interface{})
	filter["language"] = language.LangMichelson

	contracts, err := ctx.ES.GetContracts(filter)
	if err != nil {
		return err
	}

	logger.Info("Found %d contracts", len(contracts))

	var bulk []elastic.Model
	bar := progressbar.NewOptions(len(contracts), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish())

	for i := range contracts {
		bar.Add(1)

		contracts[i].Language = language.LangUnknown

		bulk = append(bulk, &contracts[i])
	}

	if err := ctx.ES.BulkUpdate(bulk); err != nil {
		logger.Errorf("ctx.ES.BulkUpdate error: %v", err)
		return err
	}

	logger.Info("Done. Total contracts: %d.", len(contracts))

	return nil
}
