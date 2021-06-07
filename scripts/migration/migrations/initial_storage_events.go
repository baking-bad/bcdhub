package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

// InitialStorageEvents -
type InitialStorageEvents struct {
	contracts map[string]string
}

// Key -
func (m *InitialStorageEvents) Key() string {
	return "execute_initial_storage"
}

// Description -
func (m *InitialStorageEvents) Description() string {
	return "execute all initial storages"
}

// Do - migrate function
func (m *InitialStorageEvents) Do(ctx *config.Context) error {
	m.contracts = make(map[string]string)
	tzips, err := ctx.TZIP.GetWithEvents(0)
	if err != nil {
		return err
	}

	logger.Info("Found %d tzips", len(tzips))

	h := metrics.New(ctx.Contracts, ctx.BigMapDiffs, ctx.Blocks, ctx.Protocols, ctx.Operations, ctx.TokenBalances, ctx.TokenMetadata, ctx.TZIP, ctx.Migrations, ctx.Storage)

	logger.Info("Execution events...")
	items := make([]models.Model, 0)
	for i := range tzips {
		logger.Info("%s...", tzips[i].Address)

		rpc, err := ctx.GetRPC(tzips[i].Network)
		if err != nil {
			return err
		}
		newModels, err := h.ExecuteInitialStorageEvent(rpc, tzips[i].Network, tzips[i].Address)
		if err != nil {
			logger.Error(err)
			continue
		}
		for i := range newModels {
			t, ok := newModels[i].(*transfer.Transfer)
			if !ok {
				items = append(items, newModels[i])
				continue
			}
			found, err := ctx.Transfers.Get(transfer.GetContext{
				Hash:    t.Hash,
				Network: t.Network,
				Counter: &t.Counter,
				Nonce:   t.Nonce,
				TokenID: &t.TokenID,
			})
			if err != nil {
				if !ctx.Storage.IsRecordNotFound(err) {
					return err
				}
			}
			if len(found.Transfers) > 0 {
				continue
			}

			items = append(items, newModels[i])
			m.contracts[t.Contract] = t.Network.String()
		}
	}

	if len(items) == 0 {
		return nil
	}

	logger.Info("Found %d models", len(items))
	return ctx.Storage.Save(items)
}

// AffectedContracts -
func (m *InitialStorageEvents) AffectedContracts() map[string]string {
	return m.contracts
}
