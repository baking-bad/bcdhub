package bigmapdiff

import (
	"context"
	"database/sql"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

func bigMapKey(q *bun.SelectQuery, keyHash string, ptr int64) *bun.SelectQuery {
	return q.Where("key_hash = ?", keyHash).Where("ptr = ?", ptr)
}

// CurrentByKey -
func (storage *Storage) Current(ctx context.Context, keyHash string, ptr int64) (data bigmapdiff.BigMapState, err error) {
	if ptr < 0 {
		err = errors.Wrapf(consts.ErrInvalidPointer, "%d", ptr)
		return
	}

	err = bigMapKey(storage.DB.NewSelect().Model(&data), keyHash, ptr).Scan(ctx)
	return
}

// GetForAddress -
func (storage *Storage) GetForAddress(ctx context.Context, address string) (response []bigmapdiff.BigMapState, err error) {
	err = core.Contract(storage.DB.NewSelect().Model(&response), address).Order("id desc").Scan(ctx)
	return
}

// GetByAddress -
func (storage *Storage) GetByAddress(ctx context.Context, address string) (response []bigmapdiff.BigMapDiff, err error) {
	err = core.Contract(storage.DB.NewSelect().Model(&response), address).Order("id desc").Scan(ctx)
	return
}

// Count -
func (storage *Storage) Count(ctx context.Context, ptr int64) (int, error) {
	return storage.DB.NewSelect().
		Model((*bigmapdiff.BigMapState)(nil)).
		Where("ptr = ?", ptr).
		Count(ctx)
}

// Previous -
func (storage *Storage) Previous(ctx context.Context, filters []bigmapdiff.BigMapDiff) ([]bigmapdiff.BigMapDiff, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	response := make([]bigmapdiff.BigMapDiff, 0)

	for i := range filters {
		var prev bigmapdiff.BigMapDiff
		if err := storage.DB.NewSelect().Model(&prev).
			Where("id < ?", filters[i].ID).
			Where("key_hash = ?", filters[i].KeyHash).
			Where("ptr = ? ", filters[i].Ptr).
			Order("id desc").Limit(1).
			Scan(ctx); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, err
		}
		response = append(response, prev)
	}

	return response, nil
}

// GetForOperation -
func (storage *Storage) GetForOperation(ctx context.Context, id int64) (response []bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.NewSelect().
		Model(&response).
		Where("operation_id = ?", id).
		Scan(ctx)
	return
}

// GetByPtrAndKeyHash -
func (storage *Storage) GetByPtrAndKeyHash(ctx context.Context, ptr int64, keyHash string, size, offset int64) ([]bigmapdiff.BigMapDiff, int64, error) {
	if ptr < 0 {
		return nil, 0, errors.Wrapf(consts.ErrInvalidPointer, "%d", ptr)
	}
	limit := storage.GetPageSize(size)

	var response []bigmapdiff.BigMapDiff

	query := storage.DB.NewSelect().Model(&response).
		Where("key_hash = ?", keyHash).
		Where("ptr = ?", ptr)
	query = core.OrderByLevelDesc(query)

	if err := query.
		Limit(limit).
		Offset(int(offset)).
		Scan(ctx); err != nil {
		return nil, 0, err
	}

	count, err := storage.DB.NewSelect().
		Model((*bigmapdiff.BigMapDiff)(nil)).
		Where("key_hash = ?", keyHash).
		Where("ptr = ?", ptr).
		Count(ctx)

	return response, int64(count), err
}

// TODO: think about remove it
// GetByPtr -
func (storage *Storage) GetByPtr(ctx context.Context, contract string, ptr int64) (response []bigmapdiff.BigMapState, err error) {
	query := storage.DB.NewSelect().
		Model(&response).
		Where("ptr = ?", ptr)
	err = core.Contract(query, contract).Scan(ctx)
	return
}

// Get -
func (storage *Storage) Get(ctx context.Context, reqCtx bigmapdiff.GetContext) ([]bigmapdiff.Bucket, error) {
	if reqCtx.Ptr != nil && *reqCtx.Ptr < 0 {
		return nil, errors.Errorf("Invalid pointer value: %d", *reqCtx.Ptr)
	}

	var bmd []bigmapdiff.Bucket
	subQuery := storage.buildGetContext(reqCtx)

	query := storage.DB.NewSelect().
		Model((*bigmapdiff.BigMapDiff)(nil)).
		ColumnExpr("*, bmd.keys_count").
		Join("inner join (?) as bmd on bmd.id = big_map_diff.id", subQuery)

	err := query.Scan(ctx, &bmd)
	return bmd, err
}

// GetStats -
func (storage *Storage) GetStats(ctx context.Context, ptr int64) (stats bigmapdiff.Stats, err error) {
	total, err := storage.DB.NewSelect().Model((*bigmapdiff.BigMapState)(nil)).
		Where("ptr = ?", ptr).
		Count(ctx)
	if err != nil {
		return stats, err
	}

	active, err := storage.DB.NewSelect().Model((*bigmapdiff.BigMapState)(nil)).
		Where("ptr = ?", ptr).
		Where("removed = false").
		Count(ctx)
	if err != nil {
		return stats, err
	}

	if err := storage.DB.NewSelect().Model((*bigmapdiff.BigMapState)(nil)).
		Column("contract").
		Where("ptr = ?", ptr).
		Limit(1).
		Scan(ctx, &stats.Contract); err != nil {
		if !storage.IsRecordNotFound(err) {
			return stats, err
		}
	}

	stats.Active = int64(active)
	stats.Total = int64(total)

	return
}

// CurrentByContract -
func (storage *Storage) CurrentByContract(ctx context.Context, contract string) (keys []bigmapdiff.BigMapState, err error) {
	err = storage.DB.NewSelect().Model(&keys).
		Where("contract = ?", contract).
		Scan(ctx)
	return
}

// StatesChangedAtLevel -
func (storage *Storage) StatesChangedAtLevel(ctx context.Context, level int64) (states []bigmapdiff.BigMapState, err error) {
	err = storage.DB.NewSelect().Model(&states).
		Where("last_update_level = ?", level).
		Scan(ctx)
	return
}

// LastDiff -
func (storage *Storage) LastDiff(ctx context.Context, ptr int64, keyHash string, skipRemoved bool) (diff bigmapdiff.BigMapDiff, err error) {
	query := storage.DB.NewSelect().Model(&diff)
	query = bigMapKey(query, keyHash, ptr)

	if skipRemoved {
		query.Where("value is not null")
	}

	err = query.Order("id desc").Limit(1).Scan(ctx)
	return
}

// Keys -
func (storage *Storage) Keys(ctx context.Context, req bigmapdiff.GetContext) (states []bigmapdiff.BigMapState, err error) {
	err = storage.buildGetContextForState(req).Scan(ctx, &states)
	return
}
