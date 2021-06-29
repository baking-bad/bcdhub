package bigmap

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

// DiffStorage -
type DiffStorage struct {
	*core.Postgres
}

// NewDiffStorage -
func NewDiffStorage(pg *core.Postgres) *DiffStorage {
	return &DiffStorage{pg}
}

func bigMap(network types.Network, ptr int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("inner join big_maps ON big_map_id = big_maps.id AND big_maps.network = ? AND big_maps.ptr = ?", network, ptr)
	}
}

func bigMapByContract(network types.Network, contract string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("inner join big_maps ON big_map_id = big_maps.id AND big_maps.network = ? AND big_maps.contract = ?", network, contract)
	}
}

func bigMapKey(network types.Network, keyHash string, ptr int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Joins("inner join big_maps ON big_map_id = big_maps.id AND big_maps.network = ? AND big_maps.ptr = ?", network, ptr).Where("key_hash = ?", keyHash)
	}
}

// GetValuesByKey -
func (storage *DiffStorage) GetValuesByKey(keyHash string) (response []bigmap.Diff, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).
		Where("key_hash = ?", keyHash).
		Group("big_map_id").
		Preload("BigMap").
		Order("level desc").
		Find(&response).
		Error
	return
}

type pointer struct {
	Network types.Network
	Ptr     int64
}

// Previous -
func (storage *DiffStorage) Previous(filters []bigmap.Diff) (response []bigmap.Diff, err error) {
	if len(filters) == 0 {
		return
	}

	pointers := map[pointer]struct{}{}

	bigMapsQuery := storage.DB.Table(models.DocBigMaps).Select("id").Where(
		storage.DB.Where("network = ? AND ptr = ?", filters[0].BigMap.Network, filters[0].BigMap.Ptr),
	)
	pointers[pointer{filters[0].BigMap.Network, filters[0].BigMap.Ptr}] = struct{}{}

	lastID := filters[0].ID
	for i := 1; i < len(filters); i++ {
		p := pointer{filters[i].BigMap.Network, filters[i].BigMap.Ptr}
		if _, ok := pointers[p]; ok {
			continue
		}

		bigMapsQuery.Or(
			storage.DB.Where("network = ? AND ptr = ?", p.Network, p.Ptr),
		)

		if lastID > filters[i].ID {
			lastID = filters[i].ID
		}
	}

	query := storage.DB.Table(models.DocBigMapDiff).
		Select("MAX(id) as id").
		Where("big_map_id IN (?)", bigMapsQuery)

	keyHashFilter := storage.DB.Where("key_hash = ?", filters[0].KeyHash)
	for i := 1; i < len(filters); i++ {
		keyHashFilter.Or(
			storage.DB.Where("key_hash = ?", filters[i].KeyHash),
		)
	}
	query.Where(keyHashFilter)

	if lastID > 0 {
		query.Where("id < ?", lastID)
	}

	query.Group("key_hash,big_map_id")

	err = storage.DB.Table(models.DocBigMapDiff).
		Where("id IN (?)", query).
		Find(&response).Error

	return
}

// GetForOperation -
func (storage *DiffStorage) GetForOperation(id int64) (response []*bigmap.Diff, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).Preload("BigMap").Where("operation_id = ?", id).Find(&response).Error
	return
}

// GetForOperations -
func (storage *DiffStorage) GetForOperations(ids ...int64) (response []bigmap.Diff, err error) {
	if len(ids) == 0 {
		return nil, nil
	}
	query := storage.DB.Table(models.DocBigMapDiff).Preload("BigMap")

	filters := storage.DB.Where(storage.DB.Where("operation_id = ?", ids[0]))
	for i := 1; i < len(ids); i++ {
		filters.Or(storage.DB.Where("operation_id = ?", ids[i]))
	}

	err = query.Where(filters).Find(&response).Error
	return
}

// GetByPtrAndKeyHash -
func (storage *DiffStorage) GetByPtrAndKeyHash(ptr int64, network types.Network, keyHash string, size, offset int64) ([]bigmap.Diff, int64, error) {
	if ptr < 0 {
		return nil, 0, errors.Wrapf(consts.ErrInvalidPointer, "%d", ptr)
	}
	limit := storage.GetPageSize(size)

	var response []bigmap.Diff
	if err := storage.DB.
		Scopes(bigMapKey(network, keyHash, ptr), core.OrderByLevelDesc).
		Preload("BigMap").
		Limit(limit).
		Offset(int(offset)).
		Debug().
		Find(&response).Error; err != nil {
		return nil, 0, err
	}

	var count int64
	err := storage.DB.Table(models.DocBigMapDiff).
		Scopes(bigMapKey(network, keyHash, ptr)).
		Count(&count).Error

	return response, count, err
}

// Get -
func (storage *DiffStorage) Get(ctx bigmap.GetContext) ([]bigmap.Bucket, error) {
	if *ctx.Ptr < 0 {
		return nil, errors.Errorf("Invalid pointer value: %d", *ctx.Ptr)
	}

	var bmd []bigmap.Bucket
	subQuery := buildGetContext(storage.DB, ctx, storage.GetPageSize(ctx.Size))

	query := storage.DB.Table(models.DocBigMapDiff).Select("*, bmd.keys_count").Joins("inner join (?) as bmd on bmd.id = big_map_diffs.id", subQuery)
	err := query.Find(&bmd).Error
	return bmd, err
}

// Last -
func (storage *DiffStorage) Last(network types.Network, ptr int64, keyHash string, skipRemoved bool) (diff bigmap.Diff, err error) {
	query := storage.DB.Table(models.DocBigMapDiff).
		Scopes(bigMapKey(network, keyHash, ptr))

	if skipRemoved {
		query.Where("value is not null")
	}

	err = query.Order("big_map_diffs.id desc").Limit(1).Scan(&diff).Error
	return
}
