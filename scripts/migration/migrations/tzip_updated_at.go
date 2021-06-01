package migrations

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// TZIPUpdatedAt - add `updated_at` column to TZIP table
type TZIPUpdatedAt struct{}

// Key -
func (m *TZIPUpdatedAt) Key() string {
	return "tzip_updated_at"
}

// Description -
func (m *TZIPUpdatedAt) Description() string {
	return "add `updated_at` column to TZIP table"
}

// Do - migrate function
func (m *TZIPUpdatedAt) Do(ctx *config.Context) error {
	model := new(tzip.TZIP)
	if err := ctx.StorageDB.DB.AutoMigrate(model); err != nil {
		return err
	}
	return ctx.StorageDB.DB.Table(models.DocTZIP).Where("id > -1").Update("updated_at", time.Now().Unix()).Error
}
