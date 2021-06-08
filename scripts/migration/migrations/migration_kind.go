package migrations

import (
	"errors"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
)

// MigrationKind -
type MigrationKind struct{}

// Key -
func (m *MigrationKind) Key() string {
	return "migration_kind"
}

// Description -
func (m *MigrationKind) Description() string {
	return "change migration kind type from string to int2"
}

// Do - migrate function
func (m *MigrationKind) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		migrator := tx.Migrator()

		model := new(migration.Migration)

		if !migrator.HasColumn(model, "kind") {
			return errors.New("Migration does not has column 'kind'")
		}

		columnTypes, err := migrator.ColumnTypes(model)
		if err != nil {
			return err
		}

		for i := range columnTypes {
			if columnTypes[i].Name() != "kind" {
				continue
			}

			if columnTypes[i].DatabaseTypeName() == "int2" {
				break
			}

			if err := migrator.RenameColumn(model, "kind", "old_kind"); err != nil {
				return err
			}

			if err := migrator.AddColumn(model, "kind"); err != nil {
				return err
			}

			for _, kind := range []types.MigrationKind{
				types.MigrationKindBootstrap, types.MigrationKindLambda, types.MigrationKindUpdate,
			} {
				if err := tx.Model(model).Where("old_kind = ?", kind.String()).Update("kind", kind).Error; err != nil {
					return err
				}
			}

			if err := migrator.DropColumn(model, "old_kind"); err != nil {
				return err
			}

			break
		}
		return nil
	})
}
