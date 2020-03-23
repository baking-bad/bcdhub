package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
)

// SetTimestampMigration - migration that set timestamp from block head to operation
type SetTimestampMigration struct{}

// Do - migrate function
func (m *SetTimestampMigration) Do(ctx *Context) error {
	for _, network := range []string{consts.Mainnet, consts.Zeronet, consts.Carthage, consts.Babylon} {
		operations, err := ctx.ES.GetAllOperations(network)
		if err != nil {
			return err
		}

		lastLevel := int64(-1)
		var lastTs time.Time
		for i, operation := range operations {
			if operation.Level == lastLevel {
				operation.Timestamp = lastTs
			} else {
				rpc, _ := ctx.GetRPC(operation.Network)
				ts, err := rpc.GetLevelTime(int(operation.Level))
				if err != nil {
					return err
				}
				operation.Timestamp = ts
			}
			if _, err := ctx.ES.UpdateDoc(elastic.DocOperations, operation.ID, operation); err != nil {
				return err
			}
			log.Printf("Done %d/%d", i, len(operations))
			lastTs = operation.Timestamp
			lastLevel = operation.Level
		}
	}
	return nil
}
