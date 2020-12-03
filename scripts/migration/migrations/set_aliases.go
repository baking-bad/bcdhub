package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
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
	h := metrics.New(ctx.ES, ctx.DB)
	aliases, err := ctx.ES.GetAliasesMap(consts.Mainnet)
	if err != nil {
		return err
	}
	logger.Info("Got %d aliases from es", len(aliases))

	updatedModels := make([]elastic.Model, 0)
	mainnetFilter := map[string]interface{}{
		"network": consts.Mainnet,
	}

	operations, err := ctx.ES.GetOperations(mainnetFilter, 0, false)
	if err != nil {
		return err
	}
	logger.Info("Got %d operations from es", len(operations))

	for i := range operations {
		if h.SetOperationAliases(aliases, &operations[i]) {
			updatedModels = append(updatedModels, &operations[i])
		}
	}

	contracts, err := ctx.ES.GetContracts(mainnetFilter)
	if err != nil {
		return err
	}

	logger.Info("Got %d contracts from es", len(contracts))

	for i := range contracts {
		if h.SetContractAlias(aliases, &contracts[i]) {
			updatedModels = append(updatedModels, &contracts[i])
		}
	}

	transfers, err := ctx.ES.GetAllTransfers(consts.Mainnet, 0)
	if err != nil {
		return err
	}

	logger.Info("Got %d transfers from es", len(transfers))

	for i := range transfers {
		if h.SetTransferAliases(aliases, &transfers[i]) {
			updatedModels = append(updatedModels, &transfers[i])
		}
	}

	logger.Info("Ready to update %d models", len(updatedModels))

	return ctx.ES.BulkUpdate(updatedModels)
}
