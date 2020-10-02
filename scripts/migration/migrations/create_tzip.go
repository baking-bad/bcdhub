package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip"
	"github.com/schollz/progressbar/v3"
)

// CreateTZIP -
type CreateTZIP struct{}

// Key -
func (m *CreateTZIP) Key() string {
	return "create_tzip"
}

// Description -
func (m *CreateTZIP) Description() string {
	return "creates tzip metadata"
}

// Do - migrate function
func (m *CreateTZIP) Do(ctx *config.Context) error {
	bmd, err := ctx.ES.GetBigMapValuesByKey(tzip.EmptyStringKey)
	if err != nil {
		return err
	}

	logger.Info("Found %d big maps with empty key", len(bmd))

	data := make([]elastic.Model, 0)
	bar := progressbar.NewOptions(len(bmd), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for i := range bmd {
		if err := bar.Add(1); err != nil {
			return err
		}

		rpc, err := ctx.GetRPC(bmd[i].Network)
		if err != nil {
			return err
		}
		parser := tzip.NewParser(ctx.ES, rpc, tzip.ParserConfig{
			IPFSGateways: ctx.Config.IPFSGateways,
		})

		t, err := parser.Parse(tzip.ParseContext{
			Address:  bmd[i].Address,
			Network:  bmd[i].Network,
			Protocol: bmd[i].Protocol,
			Pointer:  bmd[i].Ptr,
		})
		if err != nil {
			return err
		}
		if t != nil {
			data = append(data, t)
		}
	}

	return ctx.ES.BulkInsert(data)
}
