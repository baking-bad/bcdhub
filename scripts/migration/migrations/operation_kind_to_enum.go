package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
)

// OperationKindToEnum -
type OperationKindToEnum struct{}

// Key -
func (m *OperationKindToEnum) Key() string {
	return "operation_kind"
}

// Description -
func (m *OperationKindToEnum) Description() string {
	return "change operation kind type from string to int2"
}

// Do - migrate function
func (m *OperationKindToEnum) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		migrator := tx.Migrator()
		model := new(operation.Operation)

		if !migrator.HasColumn(model, "kind") {
			return nil
		}

		logger.Info().Msg("renaming 'kind' column to 'old_kind'...")
		if err := migrator.RenameColumn(model, "kind", "old_kind"); err != nil {
			return err
		}

		logger.Info().Msg("creating new 'kind' column...")
		if err := migrator.AddColumn(model, "kind"); err != nil {
			return err
		}

		logger.Info().Msg("setting 'kind' column value...")
		for _, kind := range []types.OperationKind{
			types.OperationKindOrigination, types.OperationKindOriginationNew, types.OperationKindTransaction,
		} {
			if err := tx.Model(model).Where("old_kind = ?", kind.String()).Update("kind", kind).Error; err != nil {
				return err
			}
		}
		logger.Info().Msg("removing 'old_kind' column...")

		return migrator.DropColumn(model, "old_kind")
	})
}
