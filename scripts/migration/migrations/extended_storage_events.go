package migrations

import (
	"context"
	"errors"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract_metadata"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers/stacktrace"
	transferParsers "github.com/baking-bad/bcdhub/internal/parsers/transfer"
	"github.com/go-pg/pg/v10"
)

// ExtendedStorageEvents -
type ExtendedStorageEvents struct {
	contracts map[string]types.Network
}

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
				if !ctx.Storage.IsRecordNotFound(err) {
					return err
				}
				protocol.Hash = bcd.GetCurrentProtocol()
				protocol.SymLink, err = bcd.GetProtoSymLink(protocol.Hash)
				if err != nil {
					return err
				}
			}
			rpc, err := ctx.GetRPC(tzips[i].Network)
			if err != nil {
				return err
			}
			parser, err := transferParsers.NewParser(rpc, ctx.ContractMetadata, ctx.Blocks, ctx.TokenBalances, ctx.Accounts,
				transferParsers.WithNetwork(tzips[i].Network),
				transferParsers.WithGasLimit(protocol.Constants.HardGasLimitPerOperation),
			)
			if err != nil {
				return err
			}
			script, err := ctx.Contracts.Script(tzips[i].Network, tzips[i].Address, protocol.SymLink)
			if err != nil {
				return err
			}

			for _, event := range tzips[i].Events {
				for _, impl := range event.Implementations {
					if impl.MichelsonExtendedStorageEvent == nil || impl.MichelsonExtendedStorageEvent.Empty() {
						continue
					}
					logger.Info().Msgf("%s...", tzips[i].Address)

					m.contracts[tzips[i].Address] = tzips[i].Network

					var lastID int64
					var end bool
					for !end {
						operations, err := m.getOperations(ctx, tzips[i], lastID, 10000, impl)
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

							bmd, err := ctx.BigMapDiffs.GetForOperation(op.ID)
							if err != nil {
								if !ctx.Storage.IsRecordNotFound(err) {
									return err
								}
							}
							proto, err := ctx.Cache.ProtocolByID(op.Network, op.ProtocolID)
							if err != nil {
								return err
							}

							ptrsBmd := make([]*bigmapdiff.BigMapDiff, len(bmd))
							for i := range bmd {
								ptrsBmd[i] = &bmd[i]
							}

							if err := parser.Parse(ptrsBmd, proto.Hash, &op); err != nil {
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

func (m *ExtendedStorageEvents) getOperations(ctx *config.Context, tzip contract_metadata.ContractMetadata, lastID int64, size int, impl contract_metadata.EventImplementation) ([]operation.Operation, error) {
	operations := make([]operation.Operation, 0)

	query := ctx.StorageDB.DB.Model((*operation.Operation)(nil)).
		Order("operation.id asc").
		Where("operation.network = ?", tzip.Network).
		Where("destination.address = ?", tzip.Address).
		Where("kind = ?", types.OperationKindTransaction).
		Where("status = ?", types.OperationStatusApplied).
		WhereIn("entrypoint IN (?)", impl.MichelsonExtendedStorageEvent.Entrypoints)
	if lastID > 0 {
		query.Where("operation.id > ?", lastID)
	}
	err := query.Limit(size).Relation("Destination").Relation("Source").Relation("Initiator").Select(&operations)
	return operations, err
}

// AffectedContracts -
func (m *ExtendedStorageEvents) AffectedContracts() map[string]types.Network {
	return m.contracts
}
