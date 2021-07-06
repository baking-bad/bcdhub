package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"gorm.io/gorm"
)

// ProtocolField - migration that change protocol field string -> int
type ProtocolField struct{}

// Key -
func (m *ProtocolField) Key() string {
	return "protocol"
}

// Description -
func (m *ProtocolField) Description() string {
	return " migration that change protocol field string -> int"
}

// Do - migrate function
func (m *ProtocolField) Do(ctx *config.Context) error {
	protocols, err := ctx.Protocols.GetAll()
	if err != nil {
		return err
	}
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		if err := m.migrateBlocks(tx, protocols); err != nil {
			return err
		}
		if err := m.migrateBigMapDiff(tx, protocols); err != nil {
			return err
		}
		if err := m.migrateOperations(tx, protocols); err != nil {
			return err
		}
		if err := m.migrateMigrations(tx, protocols); err != nil {
			return err
		}
		return nil
	})
}

func (m *ProtocolField) migrateBlocks(tx *gorm.DB, protocols []protocol.Protocol) error {
	logger.Info().Msg("Migrating blocks...")
	migrator := tx.Migrator()

	model := new(block.Block)

	if !migrator.HasColumn(model, "protocol_id") {
		logger.Info().Msg("Adding `protocol_id` column...")
		if err := migrator.AddColumn(model, "protocol_id"); err != nil {
			return err
		}
	}

	if migrator.HasColumn(model, "protocol") {
		logger.Info().Msg("Setting `protocol_id` value...")
		for i := range protocols {
			if err := tx.Model(model).Where("protocol = ?", protocols[i].Hash).Where("network = ?", protocols[i].Network).Update("protocol_id", protocols[i].ID).Error; err != nil {
				return err
			}
		}

		logger.Info().Msg("Removing `protocol` column...")
		return migrator.DropColumn(model, "protocol")
	}
	return nil
}

func (m *ProtocolField) migrateBigMapDiff(tx *gorm.DB, protocols []protocol.Protocol) error {
	logger.Info().Msg("Migrating bigmapdiff...")
	migrator := tx.Migrator()

	model := new(bigmapdiff.BigMapDiff)

	if !migrator.HasColumn(model, "protocol_id") {
		logger.Info().Msg("Adding `protocol_id` column...")
		if err := migrator.AddColumn(model, "protocol_id"); err != nil {
			return err
		}
	}

	if migrator.HasColumn(model, "protocol") {
		logger.Info().Msg("Setting `protocol_id` value...")
		for i := range protocols {
			if err := tx.Model(model).Where("protocol = ?", protocols[i].Hash).Where("network = ?", protocols[i].Network).Update("protocol_id", protocols[i].ID).Error; err != nil {
				return err
			}
		}

		logger.Info().Msg("Removing `protocol` column...")
		return migrator.DropColumn(model, "protocol")
	}
	return nil
}

func (m *ProtocolField) migrateOperations(tx *gorm.DB, protocols []protocol.Protocol) error {
	logger.Info().Msg("Migrating operaitons...")
	migrator := tx.Migrator()

	model := new(operation.Operation)

	if !migrator.HasColumn(model, "protocol_id") {
		logger.Info().Msg("Adding `protocol_id` column...")
		if err := migrator.AddColumn(model, "protocol_id"); err != nil {
			return err
		}
	}

	if migrator.HasColumn(model, "protocol") {
		logger.Info().Msg("Setting `protocol_id` value...")
		for i := range protocols {
			if err := tx.Model(model).Where("protocol = ?", protocols[i].Hash).Where("network = ?", protocols[i].Network).Update("protocol_id", protocols[i].ID).Error; err != nil {
				return err
			}
		}

		logger.Info().Msg("Removing `protocol` column...")
		return migrator.DropColumn(model, "protocol")
	}
	return nil
}

func (m *ProtocolField) migrateMigrations(tx *gorm.DB, protocols []protocol.Protocol) error {
	logger.Info().Msg("Migrating migrations...")
	migrator := tx.Migrator()

	model := new(migration.Migration)

	if !migrator.HasColumn(model, "protocol_id") {
		logger.Info().Msg("Adding `protocol_id` column...")
		if err := migrator.AddColumn(model, "protocol_id"); err != nil {
			return err
		}
	}

	if migrator.HasColumn(model, "protocol") {
		logger.Info().Msg("Setting `protocol_id` value...")
		for i := range protocols {
			if err := tx.Model(model).Where("protocol = ?", protocols[i].Hash).Where("network = ?", protocols[i].Network).Update("protocol_id", protocols[i].ID).Error; err != nil {
				return err
			}
		}

		logger.Info().Msg("Removing `protocol` column...")
		return migrator.DropColumn(model, "protocol")
	}

	if !migrator.HasColumn(model, "prev_protocol_id") {
		logger.Info().Msg("Adding `prev_protocol_id` column...")
		if err := migrator.AddColumn(model, "prev_protocol_id"); err != nil {
			return err
		}
	}

	if migrator.HasColumn(model, "prev_protocol") {
		logger.Info().Msg("Setting `prev_protocol_id` value...")
		for i := range protocols {
			if err := tx.Model(model).Where("prev_protocol = ?", protocols[i].Hash).Where("network = ?", protocols[i].Network).Update("prev_protocol_id", protocols[i].ID).Error; err != nil {
				return err
			}
		}

		logger.Info().Msg("Removing `prev_protocol` column...")
		return migrator.DropColumn(model, "prev_protocol")
	}

	return nil
}
