package bigmapdiff

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/consts"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
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

func bigMapKey(network types.Network, keyHash string, ptr int64) func(db *orm.Query) *orm.Query {
	return func(q *orm.Query) *orm.Query {
		return q.Where("network = ?", network).
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

	query := storage.DB.Model().Table(models.DocBigMapState)
	bigMapKey(network, keyHash, ptr)(query)
	err = query.Select(&data)
	return
}

// GetForAddress -
func (storage *Storage) GetForAddress(network types.Network, address string) (response []bigmapdiff.BigMapState, err error) {
	query := storage.DB.Model().Table(models.DocBigMapState)
	core.NetworkAndContract(network, address)(query)
	err = query.Order("id desc").Select(&response)
	return
}

// GetByAddress -
func (storage *Storage) GetByAddress(network types.Network, address string) (response []bigmapdiff.BigMapDiff, err error) {
	query := storage.DB.Model().Table(models.DocBigMapDiff)
	core.NetworkAndContract(network, address)(query)
	err = query.Order("level desc").Select(&response)
	return
}

// GetValuesByKey -
func (storage *Storage) GetValuesByKey(keyHash string) (response []bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.Model().Table(models.DocBigMapDiff).
		Where("key_hash = ?", keyHash).
		Group("network, contract, ptr").
		Order("level desc").
		Select(&response)
	return
}

// Count -
func (storage *Storage) Count(network types.Network, ptr int64) (int64, error) {
	count, err := storage.DB.Model().Table(models.DocBigMapState).
		Where("network = ?", network).
		Where("ptr = ?", ptr).
		Count()
	return int64(count), err
}

// Previous -
func (storage *Storage) Previous(filters []bigmapdiff.BigMapDiff) (response []bigmapdiff.BigMapDiff, err error) {
	if len(filters) == 0 {
		return
	}
	var lastID int64
	query := storage.DB.Model().Table(models.DocBigMapDiff).ColumnExpr("MAX(id) as id").WhereOrGroup(
		func(q *orm.Query) (*orm.Query, error) {
			for i := range filters {
				q.WhereGroup(
					func(q *orm.Query) (*orm.Query, error) {
						q.Where("key_hash = ?", filters[i].KeyHash).Where("ptr = ? ", filters[i].Ptr).Where("network = ?", filters[i].Network)
						return q, nil
					},
				)

				if lastID > filters[i].ID || lastID == 0 {
					lastID = filters[i].ID
				}
			}
			return q, nil
		},
	)

	if lastID > 0 {
		query.Where("id < ?", lastID)
	}

	query.Group("key_hash,ptr")

	err = storage.DB.Model().Table(models.DocBigMapDiff).
		Where("id IN (?)", query).
		Select(&response)

	return
}

// GetForOperation -
func (storage *Storage) GetForOperation(id int64) (response []*bigmapdiff.BigMapDiff, err error) {
	err = storage.DB.Model().Table(models.DocBigMapDiff).
		Where("operation_id = ?", id).Select(&response)
	return
}

// GetForOperations -
func (storage *Storage) GetForOperations(ids ...int64) (response []bigmapdiff.BigMapDiff, err error) {
	if len(ids) == 0 {
		return nil, nil
	}
	err = storage.DB.Model().Table(models.DocBigMapDiff).WhereIn("operation_id IN (?)", ids).Select(&response)
	return
}

// GetByPtrAndKeyHash -
func (storage *Storage) GetByPtrAndKeyHash(ptr int64, network types.Network, keyHash string, size, offset int64) ([]bigmapdiff.BigMapDiff, int64, error) {
	if ptr < 0 {
		return nil, 0, errors.Wrapf(consts.ErrInvalidPointer, "%d", ptr)
	}
	limit := storage.GetPageSize(size)

	query := storage.DB.Model().Table(models.DocBigMapDiff).
		Where("key_hash = ?", keyHash).
		Where("ptr = ?", ptr)
	query = core.Network(network)(query)
	query = core.OrderByLevelDesc(query)

	var response []bigmapdiff.BigMapDiff
	if err := query.
		Limit(limit).
		Offset(int(offset)).
		Select(&response); err != nil {
		return nil, 0, err
	}

	count, err := storage.DB.Model().Table(models.DocBigMapDiff).
		Where("network = ?", network).
		Where("key_hash = ?", keyHash).
		Where("ptr = ?", ptr).
		Count()

	return response, int64(count), err
}

// GetByPtr -
func (storage *Storage) GetByPtr(network types.Network, contract string, ptr int64) (response []bigmapdiff.BigMapState, err error) {
	query := storage.DB.Model().Table(models.DocBigMapState).Where("ptr = ?", ptr)
	core.NetworkAndContract(network, contract)(query)
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
func (storage *Storage) GetStats(network types.Network, ptr int64) (stats bigmapdiff.Stats, err error) {
	totalQuery := storage.DB.Model().Table(models.DocBigMapState).
		ColumnExpr("count(contract) as count, contract, 'total' as name").
		Where("network = ?", network).
		Where("ptr = ?", ptr).
		Group("contract")

	activeQuery := storage.DB.Model().Table(models.DocBigMapState).
		ColumnExpr("count(contract) as count, contract, 'active' as name").
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

	if _, err = storage.DB.Model().Query(&rows, "(?) union all (?)", totalQuery, activeQuery); err != nil {
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
	err = storage.DB.Model().Table(models.DocBigMapState).
		Where("network = ?", network).
		Where("contract = ?", contract).
		Select(&keys)

	return
}

// StatesChangedAfter -
func (storage *Storage) StatesChangedAfter(network types.Network, level int64) (states []bigmapdiff.BigMapState, err error) {
	err = storage.DB.Model().Table(models.DocBigMapState).
		Where("network = ?", network).
		Where("last_update_level = ?", level).
		Select(&states)
	return
}

// LastDiff -
func (storage *Storage) LastDiff(network types.Network, ptr int64, keyHash string, skipRemoved bool) (diff bigmapdiff.BigMapDiff, err error) {
	query := storage.DB.Model().Table(models.DocBigMapDiff)
	bigMapKey(network, keyHash, ptr)(query)

	if skipRemoved {
		query.Where("value is not null")
	}

	err = query.Order("id desc").Limit(1).Select(&diff)
	return
}

// Keys -
func (storage *Storage) Keys(ctx bigmapdiff.GetContext) (states []bigmapdiff.BigMapState, err error) {
	if ctx.Query == "" {
		err = storage.buildGetContextForState(ctx).Select(&states)
	} else {
		query := storage.DB.Model().ColumnExpr("bmd.*, diff.keys_count").TableExpr("(?) as diff", storage.buildGetContext(ctx)).Join("left join big_map_diffs as bmd on bmd.id  = diff.id")

		var bmd []bigmapdiff.Bucket
		if err := query.Select(&bmd); err != nil {
			return states, err
		}
		states = make([]bigmapdiff.BigMapState, len(bmd))
		for i := range bmd {
			states[i] = *bmd[i].ToState()
		}
	}
	return
}
