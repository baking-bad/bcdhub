package migrations

import (
	"errors"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	transferParser "github.com/baking-bad/bcdhub/internal/parsers/transfer"
	"github.com/schollz/progressbar/v3"
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
	tzips, err := ctx.TZIP.GetWithEvents()
	if err != nil {
		return err
	}

	logger.Info("Found %d tzips", len(tzips))

	logger.Info("Execution events...")
	for i := range tzips {
		for _, event := range tzips[i].Events {
			for _, impl := range event.Implementations {
				if impl.MichelsonParameterEvent.Empty() {
					continue
				}
				logger.Info("%s...", tzips[i].Address)

				protocol, err := ctx.Protocols.GetProtocol(tzips[i].Network, "", -1)
				if err != nil {
					return err
				}
				rpc, err := ctx.GetRPC(tzips[i].Network)
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

				inserted := make([]models.Model, 0)
				deleted := make([]models.Model, 0)
				newTransfers := make([]*transfer.Transfer, 0)

				bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
				for _, op := range operations {
					if err := bar.Add(1); err != nil {
						return err
					}

					st := stacktrace.New()
					if err := st.Fill(ctx.Operations, op); err != nil {
						return err
					}

					parser, err := transferParser.NewParser(rpc, ctx.TZIP, ctx.Blocks, ctx.Storage,
						ctx.SharePath,
						transferParser.WithNetwork(tzips[i].Network),
						transferParser.WithGasLimit(protocol.Constants.HardGasLimitPerOperation),
						transferParser.WithStackTrace(st),
					)
					if err != nil {
						return err
					}

					transfers, err := parser.Parse(op, nil)
					if err != nil {
						if errors.Is(err, noderpc.InvalidNodeResponse{}) {
							logger.Error(err)
							continue
						}
						return err
					}

					for _, t := range transfers {
						old, err := ctx.Transfers.Get(transfer.GetContext{
							Hash:    t.Hash,
							Network: t.Network,
							Counter: &t.Counter,
							Nonce:   t.Nonce,
							TokenID: t.TokenID,
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

				logger.Info("Delete %d transfers", len(deleted))
				if err := ctx.Storage.BulkDelete(deleted); err != nil {
					return err
				}

				logger.Info("Found %d transfers", len(inserted))
				if err := ctx.Storage.BulkInsert(inserted); err != nil {
					return err
				}
				if err := transferParser.UpdateTokenBalances(ctx.TokenBalances, newTransfers); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (m *ParameterEvents) getOperations(ctx *config.Context, tzip tzip.TZIP, impl tzip.EventImplementation) ([]operation.Operation, error) {
	operations := make([]operation.Operation, 0)

	for i := range impl.MichelsonParameterEvent.Entrypoints {
		ops, err := ctx.Operations.Get(map[string]interface{}{
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
