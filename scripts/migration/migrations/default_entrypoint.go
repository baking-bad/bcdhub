package migrations

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/go-pg/pg/v10"
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
	return ctx.StorageDB.DB.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		_, err := tx.Model().Table(models.DocOperations).Where("destination LIKE 'KT1%' AND parameters is null").Update("entrypoint", consts.DefaultEntrypoint)
		return err
	})
}
