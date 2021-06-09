package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"gorm.io/gorm"
)

// DropAmountStringColumns -
type DropAmountStringColumns struct{}

// Key -
func (m *DropAmountStringColumns) Key() string {
	return "drop_amount_string_columns"
}

// Description -
func (m *DropAmountStringColumns) Description() string {
	return "drop amount string columns"
}

// Do - migrate function
func (m *DropAmountStringColumns) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		migrator := tx.Migrator()
		if migrator.HasColumn(&tokenbalance.TokenBalance{}, "balance_string") {
			if err := migrator.DropColumn(&tokenbalance.TokenBalance{}, "balance_string"); err != nil {
				return err
			}
		}

		if migrator.HasColumn(&transfer.Transfer{}, "amount_string") {
			if err := migrator.DropColumn(&transfer.Transfer{}, "amount_string"); err != nil {
				return err
			}
		}

		return nil
	})
}
