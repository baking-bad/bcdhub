package migrations

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	tzipParsers "github.com/baking-bad/bcdhub/internal/parsers/contract_metadata"
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
	bmd, err := ctx.BigMapDiffs.GetValuesByKey(tzipParsers.EmptyStringKey)
	if err != nil {
		return err
	}

	logger.Info().Msgf("Found %d big maps with empty key", len(bmd))

	data := make([]models.Model, 0)
	bar := progressbar.NewOptions(len(bmd), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for i := range bmd {
		if err := bar.Add(1); err != nil {
			return err
		}

		if _, err := ctx.ContractMetadata.Get(bmd[i].Network, bmd[i].Contract); err != nil {
			if !ctx.Storage.IsRecordNotFound(err) {
				return err
			}
		} else {
			continue
		}

		rpc, err := ctx.GetRPC(bmd[i].Network)
		if err != nil {
			return err
		}
		parser := tzipParsers.NewParser(ctx.BigMapDiffs, ctx.Blocks, ctx.Contracts, ctx.Storage, rpc, tzipParsers.ParserConfig{
			IPFSGateways: ctx.Config.IPFSGateways,
		})

		proto, err := ctx.Protocols.Get(bmd[i].Network, "", bmd[i].LastUpdateLevel)
		if err != nil {
			return err
		}

		t, err := parser.Parse(tzipParsers.ParseContext{
			BigMapDiff: bigmapdiff.BigMapDiff{
				Contract:   bmd[i].Contract,
				Network:    bmd[i].Network,
				Ptr:        bmd[i].Ptr,
				Value:      bmd[i].Value,
				KeyHash:    bmd[i].KeyHash,
				ProtocolID: proto.ID,
			},
		})
		if err != nil {
			return err
		}
		if t != nil {
			data = append(data, t)
		}
	}

	return ctx.Storage.Save(context.Background(), data)
}
