package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
	"github.com/baking-bad/bcdhub/internal/parsers"
)

// Migration -
type Migration struct {
	contracts contract.Repository
}

// NewMigration -
func NewMigration(contracts contract.Repository) Migration {
	return Migration{contracts}
}

// Parse -
func (m Migration) Parse(data noderpc.Operation, operation *operation.Operation, store parsers.Store) error {
	var bmd []noderpc.BigMapDiff
	switch {
	case data.Result != nil && data.Result.BigMapDiffs != nil:
		bmd = data.Result.BigMapDiffs
	case data.Metadata != nil && data.Metadata.OperationResult != nil && data.Metadata.OperationResult.BigMapDiffs != nil:
		bmd = data.Metadata.OperationResult.BigMapDiffs
	default:
		return nil
	}

	for i := range bmd {
		if bmd[i].Action != types.BigMapActionStringUpdate || len(bmd[i].Value) == 0 {
			continue
		}

		var tree ast.UntypedAST
		if err := json.Unmarshal(bmd[i].Value, &tree); err != nil {
			return err
		}

		if len(tree) == 0 {
			continue
		}

		if tree[0].IsLambda() {
			c, err := m.contracts.Get(operation.Destination.Address)
			if err != nil {
				return err
			}
			migration := &migration.Migration{
				ContractID: c.ID,
				Level:      operation.Level,
				ProtocolID: operation.ProtocolID,
				Timestamp:  operation.Timestamp,
				Hash:       operation.Hash,
				Kind:       types.MigrationKindLambda,
			}
			store.AddMigrations(migration)
			logger.Info().Fields(migration.LogFields()).Msg("Migration detected")
			return nil
		}
	}
	return nil
}
