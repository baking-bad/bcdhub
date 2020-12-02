package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	"github.com/baking-bad/bcdhub/internal/parsers/transfer"
)

// ExtendedStorageEvents -
type ExtendedStorageEvents struct{}

// Key -
func (m *ExtendedStorageEvents) Key() string {
	return "execute_extended_storage"
}

// Description -
func (m *ExtendedStorageEvents) Description() string {
	return "execute all extended storages"
}

// Do - migrate function
func (m *ExtendedStorageEvents) Do(ctx *config.Context) error {
	tzips, err := ctx.ES.GetTZIPWithEvents()
	if err != nil {
		return err
	}

	logger.Info("Found %d tzips", len(tzips))

	logger.Info("Execution events...")
	updated := make([]elastic.Model, 0)
	for i := range tzips {
		for _, event := range tzips[i].Events {
			for _, impl := range event.Implementations {
				if impl.MichelsonExtendedStorageEvent.Empty() {
					continue
				}
				logger.Info("Execution event for %s...", tzips[i].Address)

				protocol, err := ctx.ES.GetProtocol(tzips[i].Network, "", -1)
				if err != nil {
					return err
				}
				rpc, err := ctx.GetRPC(tzips[i].Network)
				if err != nil {
					return err
				}

				parser, err := transfer.NewParser(rpc, ctx.ES,
					transfer.WithNetwork(tzips[i].Network),
					transfer.WithGasLimit(protocol.Constants.HardGasLimitPerOperation),
					transfer.WithStackTrace(stacktrace.New()),
				)
				if err != nil {
					return err
				}

				operations, err := m.getOperations(ctx, tzips[i], impl)
				if err != nil {
					return err
				}

				if len(operations) == 0 {
					continue
				}

				for _, op := range operations {
					bmd, err := ctx.ES.GetBigMapDiffsByOperationID(op.ID)
					if err != nil {
						if !elastic.IsRecordNotFound(err) {
							return err
						}
					}
					opModels := make([]elastic.Model, len(bmd))
					for j := range bmd {
						opModels[j] = bmd[j]
					}
					transfers, err := parser.Parse(op, opModels)
					if err != nil {
						return err
					}
					for _, t := range transfers {
						updated = append(updated, t)
					}
				}
			}
		}
	}

	logger.Info("Found %d transfers", len(updated))
	return ctx.ES.BulkInsert(updated)
}

func (m *ExtendedStorageEvents) getOperations(ctx *config.Context, tzip models.TZIP, impl tzip.EventImplementation) ([]models.Operation, error) {
	operations := make([]models.Operation, 0)

	for i := range impl.MichelsonExtendedStorageEvent.Entrypoints {
		ops, err := ctx.ES.GetOperations(map[string]interface{}{
			"network":     tzip.Network,
			"destination": tzip.Address,
			"kind":        consts.Transaction,
			"status":      consts.Applied,
			"entrypoint":  impl.MichelsonExtendedStorageEvent.Entrypoints[i],
		}, 0, false)
		if err != nil {
			return nil, err
		}
		operations = append(operations, ops...)
	}

	return operations, nil
}
