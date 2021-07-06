package migrations

import (
	"errors"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// TagsToInt -
type TagsToInt struct{}

// Key -
func (m *TagsToInt) Key() string {
	return "tags_to_int"
}

// Description -
func (m *TagsToInt) Description() string {
	return "change tags type from string array to int"
}

// Do - migrate function
func (m *TagsToInt) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`drop materialized view if exists head_stats;`).Error; err != nil {
			return nil
		}

		if err := m.migrate(tx, new(contract.Contract)); err != nil {
			return err
		}

		return m.migrate(tx, new(operation.Operation))
	})
}

func (m *TagsToInt) migrate(tx *gorm.DB, model models.Model) error {
	logger.Info().Msgf("migrating %s...", model.GetIndex())

	type item struct {
		ID   int64
		Tags pq.StringArray `gorm:"column:old_tags;type:text[]"`
	}

	migrator := tx.Migrator()

	if !migrator.HasColumn(model, "tags") {
		return errors.New("contract does not has column 'tags'")
	}

	columnTypes, err := migrator.ColumnTypes(model)
	if err != nil {
		return err
	}

	for i := range columnTypes {
		if columnTypes[i].Name() != "tags" {
			continue
		}

		if columnTypes[i].DatabaseTypeName() == "int8" {
			break
		}

		if err := migrator.RenameColumn(model, "tags", "old_tags"); err != nil {
			return err
		}

		if err := migrator.AddColumn(model, "tags"); err != nil {
			return err
		}

		var offset int
		limit := 10000

		end := false
		for !end {
			var items []item
			if err := tx.Table(model.GetIndex()).Where("old_tags is not null and array_length(old_tags, 1) > 0").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
				return err
			}

			for i := range items {
				tags := types.NewTags(items[i].Tags)
				if err := tx.Table(model.GetIndex()).Where("id = ?", items[i].ID).Update("tags", tags).Error; err != nil {
					return err
				}
			}

			offset += len(items)
			end = len(items) < limit
		}

		if err := migrator.DropColumn(model, "old_tags"); err != nil {
			return err
		}

		break
	}
	return nil
}
