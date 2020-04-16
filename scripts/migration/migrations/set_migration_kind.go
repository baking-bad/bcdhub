package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// SetMigrationKind - migration that sets `Kind` & `PrevProtocol` field to all contract migrations
type SetMigrationKind struct{}

// Description -
func (m *SetMigrationKind) Description() string {
	return "set `Kind` and `PrevProtocol` to contract migrations in all networks"
}

// Do - migrate function
func (m *SetMigrationKind) Do(ctx *Context) error {
	for _, network := range []string{"mainnet", "zeronet", "carthagenet", "babylonnet"} {
		migrations, err := ctx.ES.GetAllMigrations(network)
		if err != nil {
			return err
		}

		logger.Info("Found %d migrations", len(migrations))

		bar := progressbar.NewOptions(len(migrations), progressbar.OptionSetPredictTime(false))
		var bulk []elastic.BulkUpdateItem

		for i := range migrations {
			bar.Add(1)

			if migrations[i].Kind != "" {
				fmt.Print("\033[2K\r")
				logger.Warning("Already migrated.")
				return nil
			}

			if migrations[i].Level == 0 {
				migrations[i].Kind = consts.MigrationBootstrap
				migrations[i].Level = 1
				migrations[i].Protocol = "PtCJ7pwoxe8JasnHY8YonnLYjcVHmhiARPJvqcC6VfHT5s8k8sY"
			} else if migrations[i].Hash != "" {
				migrations[i].Kind = consts.MigrationLambda
			} else {
				migrations[i].Kind = consts.MigrationUpdate
				migrations[i].Level = migrations[i].Level - 1
				migrations[i].PrevProtocol = "Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd"
			}

			bulk = append(bulk, migrations[i])

			if len(bulk) == 100 || (i == len(migrations)-1 && len(bulk) > 0) {
				if err := ctx.ES.BulkUpdate("migration", bulk); err != nil {
					fmt.Print("\033[2K\r")
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
