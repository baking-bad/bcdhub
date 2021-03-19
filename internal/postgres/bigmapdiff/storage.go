package bigmapdiff

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// CurrentByKey -
func (storage *Storage) CurrentByKey(network, keyHash string, ptr int64) (data bigmapdiff.BigMapDiff, err error) {
	if ptr < 0 {
		err = errors.Wrapf(consts.ErrInvalidPointer, "%d", ptr)
		return
	}

	err = storage.DB.Table(models.DocBigMapDiff).
		Scopes(core.Network(network), core.OrderByLevelDesc).
		Where("key_hash = ?", keyHash).
		Where("ptr = ?", ptr).
		First(&data).
		Error

	return
}

// GetForAddress -
func (storage *Storage) GetForAddress(address string) (response []bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).
		Scopes(core.Address(address)).
		Group("key_hash").
		Order("level desc").
		Find(&response).
		Error
	return
}

// GetByAddress -
func (storage *Storage) GetByAddress(network, address string) (response []bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).
		Scopes(core.NetworkAndAddress(network, address)).
		Order("level desc").
		Find(&response).
		Error
	return
}

// GetValuesByKey -
func (storage *Storage) GetValuesByKey(keyHash string) (response []bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).
		Where("key_hash = ?", keyHash).
		Group("network, address, ptr").
		Order("level desc").
		Find(&response).
		Error
	return
}

// Count -
func (storage *Storage) Count(network string, ptr int64) (count int64, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).
		Scopes(core.Network(network)).
		Where("ptr = ?", ptr).
		Group("key_hash").
		Count(&count).
		Error
	return
}

// Previous -
func (storage *Storage) Previous(filters []bigmapdiff.BigMapDiff, indexedTime int64, address string) (response []bigmapdiff.BigMapDiff, err error) {
	query := storage.DB.Table(models.DocBigMapDiff).
		Scopes(core.Address(address)).
		Where("indexed_time < ?", indexedTime)

	if len(filters) > 0 {
		tx := storage.DB.Where("key_hash = ?", filters[0].KeyHash)
		for i := 1; i < len(filters); i++ {
			tx.Or("key_hash = ?", filters[i].KeyHash)
		}
		query.Where(tx)
	}

	err = query.Group("key_hash").Order("indexed_time desc").Find(&response).Error
	return response, err
}

// GetForOperation -
func (storage *Storage) GetForOperation(hash string, counter int64, nonce *int64) (response []*bigmapdiff.BigMapDiff, err error) {
	query := storage.DB.Table(models.DocBigMapDiff).
		Where("operation_hash = ?", hash).
		Where("operation_counter = ?", counter)

	if nonce == nil {
		query.Where("operation_nonce IS NULL")
	} else {
		query.Where("operation_nonce = ?", *nonce)
	}

	return response, query.Find(&response).Error
}

// GetUniqueForOperation -
func (storage *Storage) GetUniqueForOperation(hash string, counter int64, nonce *int64) (response []bigmapdiff.BigMapDiff, err error) {
	query := storage.DB.Table(models.DocBigMapDiff).
		Where("operation_hash = ?", hash).
		Where("operation_counter = ?", counter)

	if nonce == nil {
		query.Where("operation_nonce IS NULL")
	} else {
		query.Where("operation_nonce = ?", *nonce)
	}

	query.Group("key_hash, ptr").Order("indexed_time desc")

	return response, query.Find(&response).Error
}

// GetByPtrAndKeyHash -
func (storage *Storage) GetByPtrAndKeyHash(ptr int64, network, keyHash string, size, offset int64) ([]bigmapdiff.BigMapDiff, int64, error) {
	if ptr < 0 {
		return nil, 0, errors.Wrapf(consts.ErrInvalidPointer, "%d", ptr)
	}
	limit := core.GetPageSize(size)

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
func (storage *Storage) GetByPtr(address, network string, ptr int64) (response []bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).
		Scopes(core.NetworkAndAddress(network, address)).
		Where("ptr = ?", ptr).
		Group("key_hash, ptr").
		Order("indexed_time desc").
		Find(&response).
		Error

	return response, nil
}

type countResp struct {
	KeyHash   string
	KeysCount int64
}

// Get -
func (storage *Storage) Get(ctx bigmapdiff.GetContext) ([]bigmapdiff.Bucket, error) {
	if *ctx.Ptr < 0 {
		return nil, errors.Errorf("Invalid pointer value: %d", *ctx.Ptr)
	}

	var bmd []bigmapdiff.BigMapDiff
	query := storage.DB.Table(models.DocBigMapDiff)
	buildGetContext(query, ctx, true)

	if err := query.Find(&bmd).Error; err != nil {
		return nil, err
	}

	var counts []countResp
	countQuery := storage.DB.Table(models.DocBigMapDiff).Select("count(distinct(key_hash)) AS keys_count, key_hash")
	buildGetContext(countQuery, ctx, false)
	if err := countQuery.Select(&counts).Error; err != nil {
		return nil, err
	}

	result := make([]bigmapdiff.Bucket, 0)
	for i := range bmd {
		for j := range counts {
			if bmd[i].KeyHash != counts[j].KeyHash {
				continue
			}
			result = append(result, bigmapdiff.Bucket{
				BigMapDiff: bmd[i],
				Count:      counts[j].KeysCount,
			})
		}
	}
	return result, nil
}

// GetByIDs -
func (storage *Storage) GetByIDs(ids ...int64) (result []bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.Table(models.DocBigMapDiff).Order("id asc").Find(&result, ids).Error
	return
}

// GetStats -
func (storage *Storage) GetStats(network string, ptr int64) (stats bigmapdiff.Stats, err error) {
	subQuery := storage.DB.Table(models.DocBigMapDiff).
		Select("max(id) as id").
		Where("network = ?").
		Where("ptr = ?", ptr).
		Group("hash")

	err = storage.DB.Table(models.DocBigMapDiff).
		Where("id IN (?)", subQuery).
		Count(&stats.Total).Error
	if err != nil {
		return
	}

	err = storage.DB.Table(models.DocBigMapDiff).
		Where("id IN (?)", subQuery).
		Where("value is not null").
		Count(&stats.Active).Error

	if err != nil {
		return
	}

	err = storage.DB.Table(models.DocBigMapDiff).
		Select("address, network").
		Where("network = ?").
		Where("ptr = ?", ptr).
		Limit(1).
		Scan(&stats).Error

	return
}
