package bigmapdiff

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
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

func bigMapKey(keyHash string, ptr int64) func(db *orm.Query) *orm.Query {
	return func(q *orm.Query) *orm.Query {
		return q.Where("key_hash = ?", keyHash).Where("ptr = ?", ptr)
	}
}

// CurrentByKey -
func (storage *Storage) Current(keyHash string, ptr int64) (data bigmapdiff.BigMapState, err error) {
	if ptr < 0 {
		err = errors.Wrapf(consts.ErrInvalidPointer, "%d", ptr)
		return
	}

	query := storage.DB.Model().Table(models.DocBigMapState)
	bigMapKey(keyHash, ptr)(query)
	err = query.Select(&data)
	return
}

// GetForAddress -
func (storage *Storage) GetForAddress(address string) (response []bigmapdiff.BigMapState, err error) {
	query := storage.DB.Model().Table(models.DocBigMapState)
	core.Contract(address)(query)
	err = query.Order("id desc").Select(&response)
	return
}

// GetByAddress -
func (storage *Storage) GetByAddress(address string) (response []bigmapdiff.BigMapDiff, err error) {
	query := storage.DB.Model().Table(models.DocBigMapDiff)
	core.Contract(address)(query)
	err = query.Order("level desc").Select(&response)
	return
}

// GetValuesByKey -
func (storage *Storage) GetValuesByKey(keyHash string) (response []bigmapdiff.BigMapState, err error) {
	err = storage.DB.Model().Table(models.DocBigMapState).
		Where("key_hash = ?", keyHash).
		Order("last_update_level desc").
		Select(&response)
	return
}

// Count -
func (storage *Storage) Count(ptr int64) (int64, error) {
	count, err := storage.DB.Model().Table(models.DocBigMapState).
		Where("ptr = ?", ptr).
		Count()
	return int64(count), err
}

// Previous -
func (storage *Storage) Previous(filters []bigmapdiff.BigMapDiff) ([]bigmapdiff.BigMapDiff, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	response := make([]bigmapdiff.BigMapDiff, 0)

	for i := range filters {
		var prev bigmapdiff.BigMapDiff
		if err := storage.DB.Model(&prev).
			Where("id < ?", filters[i].ID).
			Where("key_hash = ?", filters[i].KeyHash).
			Where("ptr = ? ", filters[i].Ptr).
			Order("id desc").Limit(1).
			Select(); err != nil {
			if errors.Is(err, pg.ErrNoRows) {
				continue
			}
			return nil, err
		}
		response = append(response, prev)
	}

	return response, nil
}

// GetForOperation -
func (storage *Storage) GetForOperation(id int64) (response []bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.Model().Table(models.DocBigMapDiff).
		Where("operation_id = ?", id).Select(&response)
	return
}

// GetByPtrAndKeyHash -
func (storage *Storage) GetByPtrAndKeyHash(ptr int64, keyHash string, size, offset int64) ([]bigmapdiff.BigMapDiff, int64, error) {
	if ptr < 0 {
		return nil, 0, errors.Wrapf(consts.ErrInvalidPointer, "%d", ptr)
	}
	limit := storage.GetPageSize(size)

	query := storage.DB.Model().Table(models.DocBigMapDiff).
		Where("key_hash = ?", keyHash).
		Where("ptr = ?", ptr)
	query = core.OrderByLevelDesc(query)

	var response []bigmapdiff.BigMapDiff
	if err := query.
		Limit(limit).
		Offset(int(offset)).
		Select(&response); err != nil {
		return nil, 0, err
	}

	count, err := storage.DB.Model().Table(models.DocBigMapDiff).
		Where("key_hash = ?", keyHash).
		Where("ptr = ?", ptr).
		Count()

	return response, int64(count), err
}

// GetByPtr -
func (storage *Storage) GetByPtr(contract string, ptr int64) (response []bigmapdiff.BigMapState, err error) {
	query := storage.DB.Model().Table(models.DocBigMapState).Where("ptr = ?", ptr)
	core.Contract(contract)(query)
	err = query.Select(&response)
	return
}

// Get -
func (storage *Storage) Get(ctx bigmapdiff.GetContext) ([]bigmapdiff.Bucket, error) {
	if *ctx.Ptr < 0 {
		return nil, errors.Errorf("Invalid pointer value: %d", *ctx.Ptr)
	}

	var bmd []bigmapdiff.Bucket
	subQuery := storage.buildGetContext(ctx)

	query := storage.DB.Model().Table(models.DocBigMapDiff).ColumnExpr("*, bmd.keys_count").Join("inner join (?) as bmd on bmd.id = big_map_diffs.id", subQuery)
	err := query.Select(&bmd)
	return bmd, err
}

// GetStats -
func (storage *Storage) GetStats(ptr int64) (stats bigmapdiff.Stats, err error) {
	total, err := storage.DB.Model((*bigmapdiff.BigMapState)(nil)).
		Where("ptr = ?", ptr).
		Count()
	if err != nil {
		return stats, err
	}

	active, err := storage.DB.Model((*bigmapdiff.BigMapState)(nil)).
		Where("ptr = ?", ptr).
		Where("removed = false").
		Count()
	if err != nil {
		return stats, err
	}

	if err = storage.DB.Model((*bigmapdiff.BigMapState)(nil)).
		Column("contract").
		Where("ptr = ?", ptr).
		Limit(1).
		Select(&stats.Contract); err != nil {
		return
	}

	stats.Active = int64(active)
	stats.Total = int64(total)

	return
}

// CurrentByContract -
func (storage *Storage) CurrentByContract(contract string) (keys []bigmapdiff.BigMapState, err error) {
	err = storage.DB.Model().Table(models.DocBigMapState).
		Where("contract = ?", contract).
		Select(&keys)

	return
}

// StatesChangedAfter -
func (storage *Storage) StatesChangedAfter(level int64) (states []bigmapdiff.BigMapState, err error) {
	err = storage.DB.Model().Table(models.DocBigMapState).
		Where("last_update_level = ?", level).
		Select(&states)
	return
}

// LastDiff -
func (storage *Storage) LastDiff(ptr int64, keyHash string, skipRemoved bool) (diff bigmapdiff.BigMapDiff, err error) {
	query := storage.DB.Model().Table(models.DocBigMapDiff)
	bigMapKey(keyHash, ptr)(query)

	if skipRemoved {
		query.Where("value is not null")
	}

	err = query.Order("id desc").Limit(1).Select(&diff)
	return
}

// Keys -
func (storage *Storage) Keys(ctx bigmapdiff.GetContext) (states []bigmapdiff.BigMapState, err error) {
	err = storage.buildGetContextForState(ctx).Select(&states)
	return
}
