package operations

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/tidwall/gjson"
)

// Migration -
type Migration struct {
}

// NewMigration -
func NewMigration() Migration {
	return Migration{}
}

// Parse -
func (m Migration) Parse(data gjson.Result, operation *operation.Operation) (*migration.Migration, error) {
	path := "metadata.operation_result.big_map_diff"
	if !data.Get(path).Exists() {
		path = "result.big_map_diff"
		if !data.Get(path).Exists() {
			return nil, nil
		}
	}

	for _, bmd := range data.Get(path).Array() {
		if bmd.Get("action").String() != "update" || bmd.Get("value").String() == "" {
			continue
		}

		var tree ast.UntypedAST
		if err := json.UnmarshalFromString(bmd.Get("value").String(), &tree); err != nil {
			return nil, err
		}
		if len(tree) == 0 {
			continue
		}
		if tree[0].IsLambda() {
			migration := &migration.Migration{
				ID:          helpers.GenerateID(),
				IndexedTime: time.Now().UnixNano() / 1000,

				Network:   operation.Network,
				Level:     operation.Level,
				Protocol:  operation.Protocol,
				Address:   operation.Destination,
				Timestamp: operation.Timestamp,
				Hash:      operation.Hash,
				Kind:      consts.MigrationLambda,
			}
			logger.With(migration).Info("Migration detected")
			return migration, nil
		}
	}
	return nil, nil
}
