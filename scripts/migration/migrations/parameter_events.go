package migrations

import (
	"context"
	"errors"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	transferParser "github.com/baking-bad/bcdhub/internal/parsers/transfer"
	"github.com/go-pg/pg/v10"
)

// ParameterEvents -
type ParameterEvents struct {
	contracts map[string]types.Network
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
	m.contracts = make(map[string]types.Network)
	tzips, err := ctx.ContractMetadata.GetWithEvents(0)
	if err != nil {
		return err
	}

	logger.Info().Msgf("Found %d tzips", len(tzips))

	logger.Info().Msg("Execution events...")
	return ctx.StorageDB.DB.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		for i := range tzips {
			protocol, err := ctx.Protocols.Get(tzips[i].Network, "", -1)
			if err != nil {
				return err
			}
			parser, err := transferParser.NewParser(ctx.RPC, ctx.ContractMetadata, ctx.Blocks, ctx.TokenBalances, ctx.Accounts,
				transferParser.WithNetwork(tzips[i].Network),
				transferParser.WithGasLimit(protocol.Constants.HardGasLimitPerOperation),
			)
			if err != nil {
				return err
			}

			script, err := ctx.Contracts.Script(tzips[i].Network, tzips[i].Address, protocol.SymLink)
			if err != nil {
				return err
			}

			destination, err := ctx.Accounts.Get(tzips[i].Network, tzips[i].Address)
			if err != nil {
				return err
			}

			var lastID int64

			for _, event := range tzips[i].Events {
				for _, impl := range event.Implementations {
					if impl.MichelsonParameterEvent == nil || impl.MichelsonParameterEvent.Empty() {
						continue
					}
					logger.Info().Msgf("%s...", tzips[i].Address)
					m.contracts[tzips[i].Address] = tzips[i].Network

					var end bool
					for !end {
						operations, err := m.getOperations(ctx, destination.ID, lastID, 10000, impl)
						if err != nil {
							return err
						}

						for _, op := range operations {
							if lastID < op.ID {
								lastID = op.ID
							}

							op.Script, err = script.Full()
							if err != nil {
								return err
							}
							op.AST, err = ast.NewScriptWithoutCode(op.Script)
							if err != nil {
								return err
							}

							st := stacktrace.New()
							if err := st.Fill(ctx.Operations, op); err != nil {
								return err
							}
							parser.SetStackTrace(st)

							proto, err := ctx.Cache.ProtocolByID(operations[i].ProtocolID)
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
								if _, err := tx.Model((*transfer.Transfer)(nil)).
									Where("network = ?", tzips[i].Network).
									Where("token_id = ?", t.TokenID).
									Where("operation_id = ?", op.ID).
									Where("contract = ?", tzips[i].Address).
									Delete(); err != nil {
									return err
								}
							}
							if len(op.Transfers) > 0 {
								if _, err := tx.Model(&op.Transfers).Returning("id").Insert(); err != nil {
									return err
								}
							}
						}

						end = len(operations) < 10000
					}
				}
			}
		}

		return nil
	})

}

func (m *ParameterEvents) getOperations(ctx *config.Context, destinationID int64, lastID int64, size int, impl contract_metadata.EventImplementation) ([]operation.Operation, error) {
	operations := make([]operation.Operation, 0)

	query := ctx.StorageDB.DB.Model((*operation.Operation)(nil)).
		Order("operation.id asc").
		Where("destination_id = ?", destinationID).
		Where("kind = ?", types.OperationKindTransaction).
		Where("status = ?", types.OperationStatusApplied).
		WhereIn("entrypoint IN (?)", impl.MichelsonParameterEvent.Entrypoints)
	if lastID > 0 {
		query.Where("operation.id > ?", lastID)
	}
	err := query.Limit(size).Relation("Destination").Relation("Source").Relation("Initiator").Select(&operations)
	return operations, err
}

// AffectedContracts -
func (m *ParameterEvents) AffectedContracts() map[string]types.Network {
	return m.contracts
}
