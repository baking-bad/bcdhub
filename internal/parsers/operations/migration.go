package operations

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/tidwall/gjson"
)

// Migration -
type Migration struct {
	operation *operation.Operation
}

// NewMigration -
func NewMigration(operation *operation.Operation) Migration {
	return Migration{operation}
}

// Parse -
func (m Migration) Parse(data gjson.Result) *migration.Migration {
	path := "metadata.operation_result.big_map_diff"
	if !data.Get(path).Exists() {
		path = "result.big_map_diff"
		if !data.Get(path).Exists() {
			return nil
		}
	}
	for _, bmd := range data.Get(path).Array() {
		if bmd.Get("action").String() != "update" {
			continue
		}

		value := bmd.Get("value")
		if contractparser.HasLambda(value) {
			logger.Info("[%s] Migration detected: %s", m.operation.Network, m.operation.Destination)
			return &migration.Migration{
				ID:          helpers.GenerateID(),
				IndexedTime: time.Now().UnixNano() / 1000,

				Network:   m.operation.Network,
				Level:     m.operation.Level,
				Protocol:  m.operation.Protocol,
				Address:   m.operation.Destination,
				Timestamp: m.operation.Timestamp,
				Hash:      m.operation.Hash,
				Kind:      consts.MigrationLambda,
			}
		}
	}
	return nil
}
