package bigmapdiff

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

func bigMapKey(network types.Network, keyHash string, ptr int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("network = ?", network).
			Where("key_hash = ?", keyHash).
			Where("ptr = ?", ptr)
	}
}

// CurrentByKey -
func (storage *Storage) Current(network types.Network, keyHash string, ptr int64) (data bigmapdiff.BigMapState, err error) {
	if ptr < 0 {
		err = errors.Wrapf(consts.ErrInvalidPointer, "%d", ptr)
		return
	}

	err = storage.DB.Table(models.DocBigMapState).
		Scopes(bigMapKey(network, keyHash, ptr)).
		First(&data).
		Error

	return
}

// GetForAddress -
func (storage *Storage) GetForAddress(network types.Network, address string) (response []bigmapdiff.BigMapState, err error) {
	err = storage.DB.Table(models.DocBigMapState).
		Scopes(core.NetworkAndContract(network, address)).
		Order("id desc").
		Find(&response).
		Error
	return
}

// GetByAddress -
func (storage *Storage) GetByAddress(network types.Network, address string) (response []bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).
		Scopes(core.NetworkAndContract(network, address)).
		Order("level desc").
		Find(&response).
		Error
	return
}

// GetValuesByKey -
func (storage *Storage) GetValuesByKey(keyHash string) (response []bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).
		Where("key_hash = ?", keyHash).
		Group("network, contract, ptr").
		Order("level desc").
		Find(&response).
		Error
	return
}

// Count -
func (storage *Storage) Count(network types.Network, ptr int64) (count int64, err error) {
	err = storage.DB.Table(models.DocBigMapState).
		Where("network = ?", network).
		Where("ptr = ?", ptr).
		Count(&count).
		Error
	return
}

// Previous -
func (storage *Storage) Previous(filters []bigmapdiff.BigMapDiff) (response []bigmapdiff.BigMapDiff, err error) {
	if len(filters) == 0 {
		return
	}
	query := storage.DB.Table(models.DocBigMapDiff).Select("MAX(id) as id")

	tx := storage.DB.Where(
		storage.DB.Where("key_hash = ?", filters[0].KeyHash).Where("ptr = ? ", filters[0].Ptr).Where("network = ?", filters[0].Network),
	)

	lastID := filters[0].ID
	for i := 1; i < len(filters); i++ {
		tx.Or(
			storage.DB.Where("key_hash = ?", filters[i].KeyHash).Where("ptr = ? ", filters[i].Ptr).Where("network = ?", filters[i].Network),
		)

		if lastID > filters[i].ID {
			lastID = filters[i].ID
		}
	}
	query.Where(tx)

	if lastID > 0 {
		query.Where("id < ?", lastID)
	}

	query.Group("key_hash,ptr")

	err = storage.DB.Table(models.DocBigMapDiff).
		Where("id IN (?)", query).
		Find(&response).Error

	return
}

// GetForOperation -
func (storage *Storage) GetForOperation(id int64) (response []*bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).
		Where("operation_id = ?", id).Find(&response).Error
	return
}

// GetForOperations -
func (storage *Storage) GetForOperations(ids ...int64) (response []bigmapdiff.BigMapDiff, err error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := storage.DB.Table(models.DocBigMapDiff)

	filters := storage.DB.Where(storage.DB.Where("operation_id = ?", ids[0]))
	for i := 1; i < len(ids); i++ {
		filters.Or(storage.DB.Where("operation_id = ?", ids[i]))
	}

	err = query.Where(filters).Find(&response).Error
	return
}

// GetByPtrAndKeyHash -
func (storage *Storage) GetByPtrAndKeyHash(ptr int64, network types.Network, keyHash string, size, offset int64) ([]bigmapdiff.BigMapDiff, int64, error) {
	if ptr < 0 {
		return nil, 0, errors.Wrapf(consts.ErrInvalidPointer, "%d", ptr)
	}
	limit := storage.GetPageSize(size)

	var response []bigmapdiff.BigMapDiff
	if err := storage.DB.
		Scopes(core.Network(network), core.OrderByLevelDesc).
		Where("key_hash = ?", keyHash).
		Where("ptr = ?", ptr).
		Limit(limit).
		Offset(int(offset)).
		Find(&response).Error; err != nil {
		return nil, 0, err
	}

	var count int64
	err := storage.DB.Table(models.DocBigMapDiff).
		Scopes(core.Network(network)).
		Where("key_hash = ?", keyHash).
		Where("ptr = ?", ptr).
		Count(&count).Error

	return response, count, err
}

// GetByPtr -
func (storage *Storage) GetByPtr(network types.Network, contract string, ptr int64) (response []bigmapdiff.BigMapState, err error) {
	err = storage.DB.Table(models.DocBigMapState).
		Scopes(core.NetworkAndContract(network, contract)).
		Where("ptr = ?", ptr).Find(&response).Error
	return
}

// Get -
func (storage *Storage) Get(ctx bigmapdiff.GetContext) ([]bigmapdiff.Bucket, error) {
	if *ctx.Ptr < 0 {
		return nil, errors.Errorf("Invalid pointer value: %d", *ctx.Ptr)
	}

	var bmd []bigmapdiff.Bucket
	subQuery := storage.buildGetContext(ctx)

	query := storage.DB.Table(models.DocBigMapDiff).Select("*, bmd.keys_count").Joins("inner join (?) as bmd on bmd.id = big_map_diffs.id", subQuery)
	err := query.Find(&bmd).Error
	return bmd, err
}

// GetStats -
func (storage *Storage) GetStats(network types.Network, ptr int64) (stats bigmapdiff.Stats, err error) {
	totalQuery := storage.DB.Table(models.DocBigMapState).
		Select("count(contract) as count, contract, 'total' as name").
		Where("network = ?", network).
		Where("ptr = ?", ptr).
		Group("contract")

	activeQuery := storage.DB.Table(models.DocBigMapState).
		Select("count(contract) as count, contract, 'active' as name").
		Where("network = ?", network).
		Where("ptr = ?", ptr).
		Where("removed = false").
		Group("contract")

	type row struct {
		Count    int64
		Contract string
		Name     string
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
		stats.Contract = rows[i].Contract
	}

	return
}

// CurrentByContract -
func (storage *Storage) CurrentByContract(network types.Network, contract string) (keys []bigmapdiff.BigMapState, err error) {
	err = storage.DB.Table(models.DocBigMapState).
		Where("network = ?", network).
		Where("contract = ?", contract).
		Find(&keys).
		Error

	return
}

// StatesChangedAfter -
func (storage *Storage) StatesChangedAfter(network types.Network, level int64) (states []bigmapdiff.BigMapState, err error) {
	err = storage.DB.Table(models.DocBigMapState).
		Where("network = ?", network).
		Where("last_update_level > ?", level).
		Find(&states).
		Error
	return
}

// LastDiff -
func (storage *Storage) LastDiff(network types.Network, ptr int64, keyHash string, skipRemoved bool) (diff bigmapdiff.BigMapDiff, err error) {
	query := storage.DB.Table(models.DocBigMapDiff).
		Scopes(bigMapKey(network, keyHash, ptr))

	if skipRemoved {
		query.Where("value is not null")
	}

	err = query.Order("id desc").Limit(1).Scan(&diff).Error
	return
}

// Keys -
func (storage *Storage) Keys(ctx bigmapdiff.GetContext) (states []bigmapdiff.BigMapState, err error) {
	if ctx.Query == "" {
		query := storage.buildGetContextForState(ctx)
		err = query.Find(&states).Error
	} else {
		query := storage.buildGetContext(ctx)

		var bmd []bigmapdiff.Bucket
		if err := storage.DB.Table(models.DocBigMapDiff).Select("*, bmd.keys_count").Joins("inner join (?) as bmd on bmd.id = big_map_diffs.id", query).Find(&bmd).Error; err != nil {
			return states, err
		}
		states = make([]bigmapdiff.BigMapState, len(bmd))
		for i := range bmd {
			states[i] = *bmd[i].ToState()
		}
	}
	return
}
