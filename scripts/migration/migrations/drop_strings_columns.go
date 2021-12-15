package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"gorm.io/gorm"
)

// DropStringsColumns - add `updated_at` column to TZIP table
type DropStringsColumns struct{}

// Key -
func (m *DropStringsColumns) Key() string {
	return "drop_strings_columns"
}

// Description -
func (m *DropStringsColumns) Description() string {
	return "drop 'parameter_strings' and 'storage_strings' columns"
}

// Do - migrate function
func (m *DropStringsColumns) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		migrator := tx.Migrator()
		model := new(operation.Operation)
		if migrator.HasColumn(model, "parameter_strings") {
			if err := migrator.DropColumn(model, "parameter_strings"); err != nil {
				return err
			}
		} else {
			logger.Warning().Msg("operations does not contain 'parameter_strings' column")
		}

		if migrator.HasColumn(model, "storage_strings") {
			if err := migrator.DropColumn(model, "storage_strings"); err != nil {
				return err
			}
		} else {
			logger.Warning().Msg("operations does not contain 'storage_strings' column")
		}

		return nil
	})
}
