package migrations

import (
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"gorm.io/gorm"
)

// BigMapActionToEnum -
type BigMapActionToEnum struct{}

// Key -
func (m *BigMapActionToEnum) Key() string {
	return "big_map_action_to_enum"
}

// Description -
func (m *BigMapActionToEnum) Description() string {
	return "change big map action type from string to int2"
}

// Do - migrate function
func (m *BigMapActionToEnum) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		migrator := tx.Migrator()
		model := new(bigmapaction.BigMapAction)

		if !migrator.HasColumn(model, "action") {
			return nil
		}

		logger.Info("renaming 'action' column to 'old_action'...")
		if err := migrator.RenameColumn(model, "action", "old_action"); err != nil {
			return err
		}

		logger.Info("creating new 'action' column...")
		if err := migrator.AddColumn(model, "action"); err != nil {
			return err
		}

		logger.Info("setting 'action' column value...")
		for _, action := range []types.BigMapAction{
			types.BigMapActionAlloc, types.BigMapActionCopy, types.BigMapActionRemove, types.BigMapActionUpdate,
		} {
			if err := tx.Model(model).Where("old_action = ?", action.String()).Update("action", action).Error; err != nil {
				return err
			}
		}
		logger.Info("removing 'old_action' column...")

		return migrator.DropColumn(model, "old_action")
	})
}
