package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"gorm.io/gorm"
)

// NullableFields -
type NullableFields struct {
	limit int
}

// NewNullableFields -
func NewNullableFields(limit int) *NullableFields {
	return &NullableFields{limit}
}

// Key -
func (m *NullableFields) Key() string {
	return "nullable_fields"
}

// Description -
func (m *NullableFields) Description() string {
	return "set some fields to nullable type"
}

// Do - migrate function
func (m *NullableFields) Do(ctx *config.Context) error {
	if m.limit == 0 {
		m.limit = 10000
	}
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		if err := m.migrateContracts(tx); err != nil {
			return err
		}
		if err := m.migrateOperations(tx); err != nil {
			return err
		}
		if err := m.migrateTransfers(tx); err != nil {
			return err
		}
		return nil
	})
}

func (m *NullableFields) migrateContracts(tx *gorm.DB) error {
	logger.Info().Msg("processing contracts...")

	var end bool
	var offset int
	for !end {
		var contracts []contract.Contract
		if err := tx.Model(&contract.Contract{}).
			Where("delegate = ''").Or("manager = ''").
			Limit(m.limit).
			Offset(offset).
			Find(&contracts).Error; err != nil {
			return err
		}

		for i := range contracts {
			fields := make(map[string]interface{})
			if contracts[i].Manager.Str == "" && contracts[i].Manager.Valid {
				contracts[i].Manager.Valid = false
				fields["manager"] = contracts[i].Manager
			}
			if contracts[i].Delegate.Str == "" && contracts[i].Delegate.Valid {
				contracts[i].Delegate.Valid = false
				fields["delegate"] = contracts[i].Delegate
			}
			if len(fields) > 0 {
				if err := tx.Model(&contracts[i]).Updates(fields).Error; err != nil {
					return err
				}
			}
		}

		offset += len(contracts)
		end = len(contracts) < m.limit
		fmt.Printf("processed %d", offset)
	}

	return nil
}

func (m *NullableFields) migrateOperations(tx *gorm.DB) error {
	logger.Info().Msg("processing operations...")

	var end bool
	var offset int
	for !end {
		var operations []operation.Operation
		if err := tx.Model(&operation.Operation{}).
			Where("entrypoint = ''").
			Limit(m.limit).
			Offset(offset).
			Find(&operations).Error; err != nil {
			return err
		}

		for i := range operations {
			if operations[i].Entrypoint.Str == "" && operations[i].Entrypoint.Valid {
				operations[i].Entrypoint.Valid = false
				if err := tx.Model(&operations[i]).Update("entrypoint", operations[i].Entrypoint).Error; err != nil {
					return err
				}
			}
		}

		offset += len(operations)
		end = len(operations) < m.limit
		fmt.Printf("processed %d\r", offset)
	}

	return nil
}

func (m *NullableFields) migrateTransfers(tx *gorm.DB) error {
	logger.Info().Msg("processing transfers...")

	var end bool
	var offset int
	for !end {
		var transfers []transfer.Transfer
		if err := tx.Model(&transfer.Transfer{}).
			Where("parent = ''").
			Limit(m.limit).
			Offset(offset).
			Find(&transfers).Error; err != nil {
			return err
		}

		for i := range transfers {
			if transfers[i].Parent.Str == "" && transfers[i].Parent.Valid {
				transfers[i].Parent.Valid = false
				if err := tx.Model(&transfers[i]).Update("parent", transfers[i].Parent).Error; err != nil {
					return err
				}
			}
		}

		offset += len(transfers)
		end = len(transfers) < m.limit
		fmt.Printf("processed %d\r", offset)
	}

	return nil
}
