package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/schollz/progressbar/v3"
)

// SetContractMigrationsCount -
type SetContractMigrationsCount struct{}

// Description -
func (m *SetContractMigrationsCount) Description() string {
	return "set contract field `migrations_count`"
}

// Do - migrate function
func (m *SetContractMigrationsCount) Do(ctx *Context) error {
	for _, network := range []string{"mainnet", "zeronet", "carthagenet", "babylonnet"} { // TODO:
		logger.Info("Migration in %s started", network)
		migrations, err := ctx.ES.GetAllMigrations(network)
		if err != nil {
			return err
		}
		logger.Info("%d migrations are in database", len(migrations))
		bar := progressbar.NewOptions(len(migrations), progressbar.OptionSetPredictTime(false))

		for i := range migrations {
			bar.Add(1)

			if err := ctx.ES.UpdateContractMigrationsCount(migrations[i].Address, migrations[i].Network); err != nil {
				fmt.Print("\033[2K\r")
				return err
			}
		}

		fmt.Print("\033[2K\r")
		logger.Info("Migration finished in %s", network)
	}

	logger.Success("Done")
	return nil
}
