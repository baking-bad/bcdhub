package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/schollz/progressbar/v3"
)

// SetMigrationKind - migration that sets `Kind` field to all contract migrations
type SetMigrationKind struct{}

// Description -
func (m *SetMigrationKind) Description() string {
	return "set `Kind` to contract migrations in all networks"
}

// Do - migrate function
func (m *SetMigrationKind) Do(ctx *Context) error {
	for _, network := range []string{consts.Mainnet, consts.Zeronet, consts.Carthage, consts.Babylon} {
		filter := make(map[string]interface{})
		filter["network"] = network

		data, err := ctx.ES.Query([]string{"migration"}, `{"size": 10000}`)
		if err != nil {
			return err
		}

		hits := data.Get("hits.hits").Array()
		if len(hits) == 10000 {
			panic("Too many migrations, sadly, you'll have to implement chunk loading. Cheers, mate!")
		}

		logger.Info("Found %d migrations", len(hits))
		bar := progressbar.NewOptions(len(hits), progressbar.OptionSetPredictTime(false))

		var bulk []interface{}

		for i, hit := range hits {
			bar.Add(1)

			if hit.Get("_source.kind").Exists() {
				fmt.Print("\033[2K\r")
				logger.Warning("Already migrated.")
				return nil
			}

			migration := models.Migration{
				ID:          hit.Get("_id").String(),
				IndexedTime: hit.Get("_source.indexed_time").Int(),
				Protocol:    hit.Get("_source.protocol").String(),
				Hash:        hit.Get("_source.hash").String(),
				Network:     hit.Get("_source.network").String(),
				Timestamp:   hit.Get("_source.timestamp").Time().UTC(),
				Level:       hit.Get("_source.level").Int(),
				Address:     hit.Get("_source.address").String(),
			}

			if hit.Get("_source.vesting").Bool() {
				migration.Kind = consts.MigrationGenesis
			} else if len(hit.Get("_source.hash").String()) > 0 {
				migration.Kind = consts.MigrationLambda
			} else {
				migration.Kind = consts.MigrationProtocol
				migration.PrevProtocol = "Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd"
			}

			bulk = append(bulk, migration)

			if len(bulk) == 100 || (i == len(hits)-1 && len(bulk) > 0) {
				if err := ctx.ES.BulkUpdate("migration", bulk); err != nil {
					return err
				}
				bulk = bulk[:0]
			}
		}

		fmt.Print("\033[2K\r")
		logger.Info("Done.")
	}
	return nil
}
