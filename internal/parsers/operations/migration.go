package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

// Migration -
type Migration struct {
}

// NewMigration -
func NewMigration() Migration {
	return Migration{}
}

// Parse -
func (m Migration) Parse(data noderpc.Operation, operation *operation.Operation) (*migration.Migration, error) {
	var bmd []noderpc.BigMapDiff
	switch {
	case data.Result != nil && data.Result.BigMapDiffs != nil:
		bmd = data.Result.BigMapDiffs
	case data.Metadata != nil && data.Metadata.OperationResult != nil && data.Metadata.OperationResult.BigMapDiffs != nil:
		bmd = data.Metadata.OperationResult.BigMapDiffs
	default:
		return nil, nil
	}

	for i := range bmd {
		if bmd[i].Action != types.BigMapActionStringUpdate || len(bmd[i].Value) == 0 {
			continue
		}

		var tree ast.UntypedAST
		if err := json.Unmarshal(bmd[i].Value, &tree); err != nil {
			return nil, err
		}

		if len(tree) == 0 {
			continue
		}

		if tree[0].IsLambda() {
			migration := &migration.Migration{
				Network:    operation.Network,
				Level:      operation.Level,
				ProtocolID: operation.ProtocolID,
				Address:    operation.Destination,
				Timestamp:  operation.Timestamp,
				Hash:       operation.Hash,
				Kind:       types.MigrationKindLambda,
			}
			logger.With(migration).Info("Migration detected")
			return migration, nil
		}
	}
	return nil, nil
}
