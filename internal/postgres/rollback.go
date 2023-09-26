package postgres

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/uptrace/bun"
)

type Rollback struct {
	tx bun.Tx
}

func NewRollback(db *bun.DB) (Rollback, error) {
	tx, err := db.Begin()
	if err != nil {
		return Rollback{}, err
	}
	return Rollback{tx}, nil
}

func (r Rollback) Commit() error {
	return r.tx.Commit()
}

func (r Rollback) Rollback() error {
	return r.tx.Rollback()
}

func (r Rollback) DeleteAll(ctx context.Context, model any, level int64) error {
	_, err := r.tx.NewDelete().
		Model(model).
		Where("level = ?", level).
		Exec(ctx)
	return err
}

func (r Rollback) StatesChangedAtLevel(ctx context.Context, level int64) (states []bigmapdiff.BigMapState, err error) {
	err = r.tx.NewSelect().Model(&states).
		Where("last_update_level = ?", level).
		Scan(ctx)
	return
}

func (r Rollback) DeleteBigMapState(ctx context.Context, state bigmapdiff.BigMapState) error {
	_, err := r.tx.NewDelete().Model(&state).WherePK().Exec(ctx)
	return err
}

func (r Rollback) LastDiff(ctx context.Context, ptr int64, keyHash string, skipRemoved bool) (diff bigmapdiff.BigMapDiff, err error) {
	query := r.tx.NewSelect().Model(&diff).
		Where("key_hash = ?", keyHash).
		Where("ptr = ?", ptr)

	if skipRemoved {
		query.Where("value is not null")
	}

	err = query.Order("id desc").Limit(1).Scan(ctx)
	return
}

func (r Rollback) SaveBigMapState(ctx context.Context, state bigmapdiff.BigMapState) error {
	_, err := r.tx.NewUpdate().
		Column("last_update_level", "last_update_time", "removed", "value").
		Model(&state).
		WherePK().
		Exec(ctx)
	return err
}

func (r Rollback) GetOperations(ctx context.Context, level int64) (ops []operation.Operation, err error) {
	err = r.tx.NewSelect().Model(&ops).
		Where("level = ?", level).
		Scan(ctx)
	return
}

func (r Rollback) GetLastAction(ctx context.Context, addressIds ...int64) (actions []models.LastAction, err error) {
	actions = make([]models.LastAction, len(addressIds))
	for i := range addressIds {
		_, err = r.tx.NewRaw(`select max(foo.ts) as time, address from (
				(select "timestamp" as ts, source_id as address from operations where source_id = ?0 order by id desc limit 1)
				union all
				(select "timestamp" as ts, destination_id as address from operations where destination_id = ?0 order by id desc limit 1)
			) as foo
			group by address`, addressIds[i]).
			Exec(ctx, &actions[i])
		if err != nil {
			return nil, err
		}
	}
	return
}

func (r Rollback) UpdateAccountStats(ctx context.Context, addressId int64, lastAction time.Time, operationsCount int64) error {
	_, err := r.tx.NewUpdate().Model((*account.Account)(nil)).
		Where("id = ?", addressId).
		Set("operations_count = operations_count - ?", operationsCount).
		Set("last_action = ?", lastAction).
		Exec(ctx)
	return err
}

func (r Rollback) GlobalConstants(ctx context.Context, level int64) (constants []contract.GlobalConstant, err error) {
	err = r.tx.NewSelect().Model(&constants).
		Where("level = ?", level).
		Scan(ctx)
	return
}

func (r Rollback) Scripts(ctx context.Context, level int64) (scripts []contract.Script, err error) {
	err = r.tx.NewSelect().Model(&scripts).
		Where("level = ?", level).
		Scan(ctx)
	return
}

func (r Rollback) DeleteScriptsConstants(ctx context.Context, scriptIds []int64, constantsIds []int64) error {
	if len(scriptIds) == 0 && len(constantsIds) == 0 {
		return nil
	}

	query := r.tx.NewDelete().
		Model((*contract.ScriptConstants)(nil)).
		Where("script_id IN (?)", bun.In(scriptIds)).
		WhereOr("global_constant_id IN (?)", bun.In(constantsIds))

	_, err := query.Exec(ctx)
	return err
}

func (r Rollback) Protocols(ctx context.Context, level int64) error {
	result, err := r.tx.NewDelete().Model((*protocol.Protocol)(nil)).Where("start_level >= ?", level).Exec(ctx)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	_, err = r.tx.NewUpdate().
		Model((*protocol.Protocol)(nil)).
		Where("start_level < ?", level).
		Set("end_level = 0").
		Exec(ctx)
	return err
}
