package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/ast"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/tidwall/gjson"
)

// TODO: relocate to storage? or metrics?

// Migration -
type Migration struct {
	operation *operation.Operation
	shareDir  string
}

// NewMigration -
func NewMigration(operation *operation.Operation, shareDir string) Migration {
	return Migration{operation, shareDir}
}

// Parse -
func (m Migration) Parse(data gjson.Result) (*migration.Migration, error) {
	path := "metadata.operation_result.big_map_diff"
	if !data.Get(path).Exists() {
		path = "result.big_map_diff"
		if !data.Get(path).Exists() {
			return nil, nil
		}
	}
	for _, bmd := range data.Get(path).Array() {
		if bmd.Get("action").String() != "update" {
			continue
		}

		value := bmd.Get("value")
		if value.Raw == "" {
			continue
		}

		var tree ast.UntypedAST
		if err := json.UnmarshalFromString(value.String(), &tree); err != nil {
			return nil, err
		}

		// if ast.HasPrimitive(tree, consts.LAMBDA) {
		// 	migration := &migration.Migration{
		// 		ID:          helpers.GenerateID(),
		// 		IndexedTime: time.Now().UnixNano() / 1000,

		// 		Network:   m.operation.Network,
		// 		Level:     m.operation.Level,
		// 		Protocol:  m.operation.Protocol,
		// 		Address:   m.operation.Destination,
		// 		Timestamp: m.operation.Timestamp,
		// 		Hash:      m.operation.Hash,
		// 		Kind:      consts.MigrationLambda,
		// 	}
		// 	logger.With(migration).Info("Migration detected")
		// 	return migration, nil
		// }
	}
	return nil, nil
}
