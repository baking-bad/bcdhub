package main

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/pkg/errors"
)

func getMigrations(ids []string) error {
	migrations := make([]migration.Migration, 0)
	if err := ctx.ES.GetByIDs(&migrations, ids...); err != nil {
		return errors.Errorf("[getMigrations] Find migration error for IDs %v: %s", ids, err)
	}

	for i := range migrations {
		if err := parseMigration(migrations[i]); err != nil {
			return errors.Errorf("[getMigrations] Compute error message: %s", err)
		}
		logger.With(&migrations[i]).Info("Migration is processed")
	}

	return nil
}

func parseMigration(migration migration.Migration) error {
	return ctx.ES.UpdateContractMigrationsCount(migration.Address, migration.Network)
}
