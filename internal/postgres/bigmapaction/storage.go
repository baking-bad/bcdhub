package bigmapaction

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// Get -
func (storage *Storage) Get(ptr int64, network string) (actions []bigmapaction.BigMapAction, err error) {
	err = storage.DB.Table(models.DocBigMapActions).
		Where(
			storage.DB.Where("network = ?", network).
				Where(
					storage.DB.Where("source_ptr = ?", ptr).Or("destination_ptr = ?", ptr),
				),
		).
		Order("id DESC").
		Find(&actions).Error
	return
}
