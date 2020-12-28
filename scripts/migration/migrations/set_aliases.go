package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
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
	h := metrics.New(ctx.Contracts, ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.Schema, ctx.TokenBalances, ctx.TZIP, ctx.Migrations, ctx.Storage, ctx.Bulk, ctx.DB)

	updatedModels := make([]elastic.Model, 0)
	for i := range ctx.Config.Scripts.Networks {
		logger.Info("Receiving aliases for %s...", ctx.Config.Scripts.Networks[i])

		aliases, err := ctx.ES.GetAliasesMap(ctx.Config.Scripts.Networks[i])
		if err != nil {
			if elastic.IsRecordNotFound(err) {
				continue
			}
			return err
		}
		logger.Info("Received %d aliases", len(aliases))

		if len(aliases) == 0 {
			continue
		}

		networkFilter := map[string]interface{}{
			"network": ctx.Config.Scripts.Networks[i],
		}
		contracts, err := ctx.Contracts.GetMany(networkFilter)
		if err != nil {
			return err
		}

		operations, err := ctx.ES.GetOperations(networkFilter, 0, false)
		if err != nil {
			return err
		}
		logger.Info("Got %d operations", len(operations))

		for i := range operations {
			if flag, err := h.SetOperationAliases(&operations[i]); flag {
				updatedModels = append(updatedModels, &operations[i])
			} else if err != nil {
				return err
			}
		}

		contracts, err := ctx.ES.GetContracts(networkFilter)
		if err != nil {
			return err
		}

		logger.Info("Got %d contracts", len(contracts))

		for i := range contracts {
			if flag, err := h.SetContractAlias(&contracts[i]); flag {
				updatedModels = append(updatedModels, &contracts[i])
			} else if err != nil {
				return err
			}
		}

		transfers, err := ctx.Transfers.GetAll(ctx.Config.Scripts.Networks[i], 0)
		if err != nil {
			return err
		}

		logger.Info("Got %d transfers", len(transfers))

		for i := range transfers {
			if flag, err := h.SetTransferAliases(&transfers[i]); flag {
				updatedModels = append(updatedModels, &transfers[i])
			} else if err != nil {
				return err
			}
		}
	}

	logger.Info("Updating %d models...", len(updatedModels))

	return ctx.Bulk.Update(updatedModels)
}
