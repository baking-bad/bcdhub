package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"gorm.io/gorm"
)

// DropAliasesColumns -
type DropAliasesColumns struct{}

// Key -
func (m *DropAliasesColumns) Key() string {
	return "drop_alias_columns"
}

// Description -
func (m *DropAliasesColumns) Description() string {
	return "drop alias columns"
}

// Do - migrate function
func (m *DropAliasesColumns) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		migrator := tx.Migrator()
		if migrator.HasColumn(&contract.Contract{}, "alias") {
			if err := migrator.DropColumn(&contract.Contract{}, "alias"); err != nil {
				return err
			}
		}

		if migrator.HasColumn(&contract.Contract{}, "delegate_alias") {
			if err := migrator.DropColumn(&contract.Contract{}, "delegate_alias"); err != nil {
				return err
			}
		}

		if migrator.HasColumn(&operation.Operation{}, "source_alias") {
			if err := migrator.DropColumn(&operation.Operation{}, "source_alias"); err != nil {
				return err
			}
		}

		if migrator.HasColumn(&operation.Operation{}, "destination_alias") {
			if err := migrator.DropColumn(&operation.Operation{}, "destination_alias"); err != nil {
				return err
			}
		}

		if migrator.HasColumn(&operation.Operation{}, "delegate_alias") {
			if err := migrator.DropColumn(&operation.Operation{}, "delegate_alias"); err != nil {
				return err
			}
		}

		return nil
	})
}
