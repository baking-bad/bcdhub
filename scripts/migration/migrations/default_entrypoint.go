package migrations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models"
	"gorm.io/gorm"
)

// DefaultEntrypoint - set entrypoint `default` to contract calls with empty parameters
type DefaultEntrypoint struct{}

// Key -
func (m *DefaultEntrypoint) Key() string {
	return "default_entrypoint"
}

// Description -
func (m *DefaultEntrypoint) Description() string {
	return "set entrypoint `default` to contract calls with empty parameters"
}

// Do - migrate function
func (m *DefaultEntrypoint) Do(ctx *config.Context) error {
	return ctx.StorageDB.DB.Transaction(func(tx *gorm.DB) error {
		return tx.Table(models.DocOperations).Where("destination LIKE 'KT1%' AND parameters is null").Update("entrypoint", consts.DefaultEntrypoint).Error
	})
}
