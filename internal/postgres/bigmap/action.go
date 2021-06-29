package bigmap

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
)

// ActionStorage -
type ActionStorage struct {
	*core.Postgres
}

// NewActionStorage -
func NewActionStorage(pg *core.Postgres) *ActionStorage {
	return &ActionStorage{pg}
}

// Get -
func (storage *ActionStorage) Get(network types.Network, ptr int64) (actions []bigmap.Action, err error) {
	err = storage.DB.Table(models.DocBigMapActions).
		Joins("left join big_maps as source on source.id = source_id").
		Joins("left join big_maps as destination on destination.id = destination_id").
		Preload("Source").Preload("Destination").
		Where(
			storage.DB.Where("source.network = ?", network).
				Where(
					storage.DB.Where("source.ptr = ?", ptr).Or("destination.ptr = ?", ptr),
				),
		).
		Order("big_map_actions.id DESC").
		Find(&actions).Error
	return
}
