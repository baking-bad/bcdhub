package main

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/pkg/errors"
)

func getMigrations(ids []int64) error {
	migrations, err := ctx.Migrations.GetByIDs(ids...)
	if err != nil {
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
	return ctx.Contracts.UpdateMigrationsCount(migration.Address, migration.Network)
}
