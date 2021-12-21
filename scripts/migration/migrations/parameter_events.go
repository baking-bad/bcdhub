package migrations

import (
	"context"
	"errors"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/fetch"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
	tzips, err := ctx.TZIP.GetWithEvents(0)
	if err != nil {
		return err
	}

	logger.Info().Msgf("Found %d tzips", len(tzips))

	logger.Info().Msg("Execution events...")
	for i := range tzips {
		for _, event := range tzips[i].Events {
			for _, impl := range event.Implementations {
				if impl.MichelsonParameterEvent.Empty() {
					continue
				}
				logger.Info().Msgf("%s...", tzips[i].Address)

				protocol, err := ctx.Protocols.Get(tzips[i].Network, "", -1)
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

				script, err := fetch.ContractBySymLink(tzips[i].Network, tzips[i].Address, protocol.SymLink, ctx.SharePath)
				if err != nil {
					return err
				}

				inserted := make([]models.Model, 0)
				deleted := make([]models.Model, 0)
				newTransfers := make([]*transfer.Transfer, 0)

				bar := progressbar.NewOptions(len(operations), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
				for _, op := range operations {
					if err := bar.Add(1); err != nil {
						return err
					}
					op.Script = script
					tree, err := ast.NewScriptWithoutCode(script)
					if err != nil {
						return err
					}
					op.AST = tree

					st := stacktrace.New()
					if err := st.Fill(ctx.Operations, op); err != nil {
						return err
					}

					parser, err := transferParser.NewParser(rpc, ctx.TZIP, ctx.Blocks, ctx.TokenBalances,
						ctx.SharePath,
						transferParser.WithNetwork(tzips[i].Network),
						transferParser.WithGasLimit(protocol.Constants.HardGasLimitPerOperation),
						transferParser.WithStackTrace(st),
					)
					if err != nil {
						return err
					}

					proto, err := ctx.CachedProtocolByID(operations[i].Network, operations[i].ProtocolID)
					if err != nil {
						return err
					}
					if err := parser.Parse(nil, proto.Hash, &op); err != nil {
						if errors.Is(err, noderpc.InvalidNodeResponse{}) {
							logger.Err(err)
							continue
						}
						return err
					}

					for _, t := range op.Transfers {
						old, err := ctx.Transfers.Get(transfer.GetContext{
							Network:     t.Network,
							TokenID:     &t.TokenID,
							OperationID: &op.ID,
						})
						if err != nil {
							return err
						}
						for j := range old.Transfers {
							deleted = append(deleted, &old.Transfers[j])
							m.contracts[old.Transfers[j].Contract] = old.Transfers[j].Network.String()
						}
						inserted = append(inserted, t)
						newTransfers = append(newTransfers, t)
						m.contracts[t.Contract] = t.Network.String()
					}
				}

				logger.Info().Msgf("Delete %d transfers", len(deleted))
				if err := ctx.Storage.BulkDelete(context.Background(), deleted); err != nil {
					return err
				}

				logger.Info().Msgf("Found %d transfers", len(inserted))
				bu := transferParser.UpdateTokenBalances(newTransfers)
				for i := range bu {
					inserted = append(inserted, bu[i])
				}

				if err := ctx.Storage.Save(context.Background(), inserted); err != nil {
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
			"kind":        types.OperationKindTransaction,
			"status":      types.OperationStatusApplied,
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
