package migrations

import (
	"errors"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/parsers/transfer"
)

// ParameterEvents -
type ParameterEvents struct {
	contracts map[string]string
}

// Key -
func (m *ParameterEvents) Key() string {
	return "execute_parameter_events"
}

// Description -
func (m *ParameterEvents) Description() string {
	return "execute all parameter events"
}

// Do - migrate function
func (m *ParameterEvents) Do(ctx *config.Context) error {
	m.contracts = make(map[string]string)
	tzips, err := ctx.ES.GetTZIPWithEvents()
	if err != nil {
		return err
	}

	logger.Info("Found %d tzips", len(tzips))

	logger.Info("Execution events...")
	inserted := make([]elastic.Model, 0)
	deleted := make([]elastic.Model, 0)
	newTransfers := make([]*models.Transfer, 0)
	for i := range tzips {
		for _, event := range tzips[i].Events {
			for _, impl := range event.Implementations {
				if impl.MichelsonParameterEvent.Empty() {
					continue
				}
				logger.Info("%s...", tzips[i].Address)

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
					transfers, err := parser.Parse(op, nil)
					if err != nil {
						if errors.Is(err, events.ErrNodeReturn) {
							logger.Error(err)
							continue
						}
						return err
					}

					for _, t := range transfers {
						old, err := ctx.ES.GetTransfers(elastic.GetTransfersContext{
							Hash:    t.Hash,
							Network: t.Network,
							Counter: &t.Counter,
							Nonce:   t.Nonce,
							TokenID: -1,
						})
						if err != nil {
							return err
						}
						for j := range old.Transfers {
							deleted = append(deleted, &old.Transfers[j])
							m.contracts[old.Transfers[j].Contract] = old.Transfers[j].Network
						}
						inserted = append(inserted, t)
						newTransfers = append(newTransfers, t)
						m.contracts[t.Contract] = t.Network
					}
				}
			}
		}
	}
	logger.Info("Delete %d transfers", len(deleted))
	if err := ctx.ES.BulkDelete(deleted); err != nil {
		return err
	}

	logger.Info("Found %d transfers", len(inserted))
	if err := ctx.ES.BulkInsert(inserted); err != nil {
		return err
	}
	return elastic.CreateTokenBalanceUpdates(ctx.ES, newTransfers)
}

func (m *ParameterEvents) getOperations(ctx *config.Context, tzip models.TZIP, impl tzip.EventImplementation) ([]models.Operation, error) {
	operations := make([]models.Operation, 0)

	for i := range impl.MichelsonParameterEvent.Entrypoints {
		ops, err := ctx.ES.GetOperations(map[string]interface{}{
			"network":     tzip.Network,
			"destination": tzip.Address,
			"kind":        consts.Transaction,
			"status":      consts.Applied,
			"entrypoint":  impl.MichelsonParameterEvent.Entrypoints[i],
		}, 0, false)
		if err != nil {
			return nil, err
		}
		operations = append(operations, ops...)
	}

	return operations, nil
}

// AffectedContracts -
func (m *ParameterEvents) AffectedContracts() map[string]string {
	return m.contracts
}
