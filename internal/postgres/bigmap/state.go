package bigmap

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
)

// StateStorage -
type StateStorage struct {
	*core.Postgres
}

// NewStateStorage -
func NewStateStorage(pg *core.Postgres) *StateStorage {
	return &StateStorage{pg}
}

// CurrentByKey -
func (storage *StateStorage) Current(network types.Network, keyHash string, ptr int64) (data bigmap.State, err error) {
	if ptr < 0 {
		err = errors.Wrapf(consts.ErrInvalidPointer, "%d", ptr)
		return
	}

	err = storage.DB.Table(models.DocBigMapState).
		Scopes(bigMapKey(network, keyHash, ptr)).
		Preload("BigMap").
		First(&data).
		Error

	return
}

// GetForAddress -
func (storage *StateStorage) GetForAddress(network types.Network, contract string) (response []bigmap.State, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).
		Scopes(bigMapByContract(network, contract)).
		Preload("BigMap").
		Order("level desc").
		Find(&response).
		Error
	return
}

// GetByPtr -
func (storage *StateStorage) GetByPtr(network types.Network, contract string, ptr int64) (response []bigmap.State, err error) {
	err = storage.DB.Table(models.DocBigMapState).
		Scopes(bigMapByContract(network, contract)).
		Preload("BigMap").
		Where("big_maps.ptr = ?", ptr).Find(&response).Error
	return
}

// GetStats -
func (storage *StateStorage) GetStats(network types.Network, ptr int64) (stats bigmap.Stats, err error) {
	totalQuery := storage.DB.Table(models.DocBigMapState).
		Select("count(contract) as count, 'total' as name").
		Scopes(bigMap(network, ptr)).
		Group("contract")

	activeQuery := storage.DB.Table(models.DocBigMapState).
		Select("count(contract) as count, 'active' as name").
		Scopes(bigMap(network, ptr)).
		Where("removed = false").
		Group("contract")

	type row struct {
		Count int64
		Name  string
	}
	var rows []row

	if err = storage.DB.
		Raw("(?) union all (?)", totalQuery, activeQuery).
		Scan(&rows).
		Error; err != nil {
		return
	}

	for i := range rows {
		switch rows[i].Name {
		case "active":
			stats.Active = rows[i].Count
		case "total":
			stats.Total = rows[i].Count
		}
	}

	return
}

// ChangedAfter -
func (storage *StateStorage) ChangedAfter(network types.Network, level int64) (states []bigmap.State, err error) {
	err = storage.DB.Table(models.DocBigMapState).
		Joins("left join big_maps on big_maps.id = big_map_id and big_maps.network = ?", network).
		Where("last_update_level > ?", level).
		Find(&states).
		Error
	return
}

// Keys -
func (storage *StateStorage) Keys(ctx bigmap.GetContext) (states []bigmap.State, err error) {
	if ctx.Query == "" {
		query := buildGetContextForState(storage.DB, ctx, storage.GetPageSize(ctx.Size))
		err = query.Find(&states).Error
	} else {
		query := buildGetContext(storage.DB, ctx, storage.GetPageSize(ctx.Size))

		var bmd []bigmap.Bucket
		if err := storage.DB.Table(models.DocBigMapDiff).Preload("BigMap").Select("*, bmd.keys_count").Joins("inner join (?) as bmd on bmd.id = big_map_diffs.id", query).Find(&bmd).Error; err != nil {
			return states, err
		}
		states = make([]bigmap.State, len(bmd))
		for i := range bmd {
			states[i] = *bmd[i].ToState()
		}
	}
	return
}

// Count -
func (storage *StateStorage) Count(network types.Network, ptr int64) (count int64, err error) {
	err = storage.DB.Table(models.DocBigMapState).
		Scopes(bigMap(network, ptr)).
		Count(&count).
		Error
	return
}
