package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
)

// SetAliases - migration that set aliases for operations, contracts and transfers
type SetAliases struct{}

// Key -
func (m *SetAliases) Key() string {
	return "set_aliases"
}

// Description -
func (m *SetAliases) Description() string {
	return "set aliases for operations, contracts and transfers"
}

// Do - migrate function
func (m *SetAliases) Do(ctx *config.Context) error {
	h := metrics.New(ctx.Contracts, ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.Schema, ctx.TokenBalances, ctx.TokenMetadata, ctx.TZIP, ctx.Migrations, ctx.Storage, ctx.DB)

	updatedModels := make([]models.Model, 0)
	logger.Info("Receiving aliases for %s...", consts.Mainnet)

	aliases, err := ctx.TZIP.GetAliasesMap(consts.Mainnet)
	if err != nil {
		return err
	}
	logger.Info("Received %d aliases", len(aliases))

	if len(aliases) == 0 {
		return nil
	}

	addresses := make([]string, 0, len(aliases))
	for address := range aliases {
		addresses = append(addresses, address)
	}

	for _, field := range []string{"source.or", "destination.or", "delegate.or"} {
		filter := map[string]interface{}{
			"network": consts.Mainnet,
			field:     addresses,
		}

		operations, err := ctx.Operations.Get(filter, 0, false)
		if err != nil {
			return err
		}

		logger.Info("Got %d operations for %s", len(operations), field)

		for i := range operations {
			if flag, err := h.SetOperationAliases(&operations[i]); flag {
				updatedModels = append(updatedModels, &operations[i])
			} else if err != nil {
				return err
			}
		}
	}

	for _, field := range []string{"address.or", "delegate.or"} {
		filter := map[string]interface{}{
			"network": consts.Mainnet,
			field:     addresses,
		}

		contracts, err := ctx.Contracts.GetMany(filter)
		if err != nil {
			return err
		}

		logger.Info("Got %d contracts for %s", len(contracts), field)

		for i := range contracts {
			if flag, err := h.SetContractAlias(&contracts[i]); flag {
				updatedModels = append(updatedModels, &contracts[i])
			} else if err != nil {
				return err
			}
		}
	}

	logger.Info("Updating %d models...", len(updatedModels))

	return ctx.Storage.BulkUpdate(updatedModels)
}
