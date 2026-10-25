package bigmapdiff

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/pkg/errors"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
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

// Count -
func (storage *Storage) Count(ctx context.Context, ptr int64) (int, error) {
	return storage.DB.NewSelect().
		Model((*bigmapdiff.BigMapState)(nil)).
		Where("ptr = ?", ptr).
		Count(ctx)
}

// Previous - returns the latest diff preceding each of `filters` for the same key and pointer in a single query. Filters without a previous diff are skipped. Result order follows `filters`.
func (storage *Storage) Previous(ctx context.Context, filters []bigmapdiff.BigMapDiff) ([]bigmapdiff.BigMapDiff, error) {
	if len(filters) == 0 {
		return nil, nil
	}

	ids := make([]int64, len(filters))
	keyHashes := make([]string, len(filters))
	ptrs := make([]int64, len(filters))
	for i := range filters {
		ids[i] = filters[i].ID
		keyHashes[i] = filters[i].KeyHash
		ptrs[i] = filters[i].Ptr
	}

	prev := storage.DB.NewSelect().
		Model((*bigmapdiff.BigMapDiff)(nil)).
		Where("big_map_diff.key_hash = f.key_hash").
		Where("big_map_diff.ptr = f.ptr").
		Where("big_map_diff.id < f.id").
		Order("big_map_diff.id desc").
		Limit(1)

	response := make([]bigmapdiff.BigMapDiff, 0, len(filters))
	err := storage.DB.NewSelect().
		TableExpr(
			"unnest(?::bigint[], ?::text[], ?::bigint[]) with ordinality as f(id, key_hash, ptr, ord)",
			pgdialect.Array(ids), pgdialect.Array(keyHashes), pgdialect.Array(ptrs),
		).
		Join("join lateral (?) as prev on true", prev).
		ColumnExpr("prev.*").
		OrderExpr("f.ord").
		Scan(ctx, &response)
	return response, err
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

// Keys -
func (storage *Storage) Keys(ctx context.Context, req bigmapdiff.GetContext) (states []bigmapdiff.BigMapState, err error) {
	err = storage.buildGetContextForState(req).Scan(ctx, &states)
	return
}
