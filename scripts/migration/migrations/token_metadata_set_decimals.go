package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/parsers/tzip/tokens"
)

// TokenMetadataSetDecimals - migration that set decimals in token metadata for all metadata with empty decimals
type TokenMetadataSetDecimals struct{}

// Key -
func (m *TokenMetadataSetDecimals) Key() string {
	return "token_metadata_set_decimals"
}

// Description -
func (m *TokenMetadataSetDecimals) Description() string {
	return "set decimals in token metadata for all metadata with empty decimals"
}

// Do - migrate function
func (m *TokenMetadataSetDecimals) Do(ctx *config.Context) error {
	var updates []models.Model

	for _, network := range ctx.Config.Scripts.Networks {
		logger.Info("Work with %swork", network)
		rpc, err := ctx.GetRPC(network)
		if err != nil {
			return err
		}

		parser := tokens.NewParser(ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Storage, rpc, ctx.SharePath, network, ctx.Config.IPFSGateways...)

		logger.Info("Receiving token metadata....")
		tokenMetadata, err := ctx.TokenMetadata.Get(tokenmetadata.GetContext{
			TokenID: -1,
			Network: network,
		})
		if err != nil {
			return err
		}

		logger.Info("Received %d metadata....", len(tokenMetadata))

		if len(tokenMetadata) == 0 {
			continue
		}

		for i, tm := range tokenMetadata {
			if tm.Decimals != nil {
				continue
			}

			parsedTm, err := parser.Parse(tm.Contract, 0)
			if err != nil {
				return err
			}

			for j, ptm := range parsedTm {
				if ptm.Decimals != nil {
					logger.Info("Found: contract=%s decimals=%v", ptm.Contract, *ptm.Decimals)
					tokenMetadata[i].Decimals = parsedTm[j].Decimals
					updates = append(updates, &tokenMetadata[i])
				}
			}
		}
	}

	logger.Info("Total updates: %d", len(updates))

	return ctx.Storage.BulkUpdate(updates)
}
