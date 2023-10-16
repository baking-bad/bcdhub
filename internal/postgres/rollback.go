package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/stats"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
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

func (r Rollback) DeleteAll(ctx context.Context, model any, level int64) (int, error) {
	result, err := r.tx.NewDelete().
		Model(model).
		Where("level = ?", level).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(count), nil
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
	actions = make([]models.LastAction, 0)
	for i := range addressIds {
		var action models.LastAction
		_, err = r.tx.NewRaw(`select max(foo.ts) as time, address from (
				(select "timestamp" as ts, source_id as address from operations where source_id = ?0 order by timestamp desc limit 1)
				union all
				(select "timestamp" as ts, destination_id as address from operations where destination_id = ?0 order by timestamp desc limit 1)
			) as foo
			group by address`, addressIds[i]).
			Exec(ctx, &action)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = nil
				continue
			}
			return nil, err
		}
		actions = append(actions, action)
	}
	return
}

func (r Rollback) UpdateAccountStats(ctx context.Context, account account.Account) error {
	_, err := r.tx.NewUpdate().Model(&account).
		Where("id = ?id").
		Set("operations_count = operations_count - ?operations_count").
		Set("migrations_count = migrations_count - ?migrations_count").
		Set("events_count = events_count - ?events_count").
		Set("ticket_updates_count = ticket_updates_count - ?ticket_updates_count").
		Set("last_action = ?last_action").
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

func (r Rollback) UpdateStats(ctx context.Context, stats stats.Stats) error {
	_, err := r.tx.NewUpdate().
		Model(&stats).
		Where("id = ?id").
		Set("contracts_count = ?contracts_count").
		Set("operations_count = ?operations_count").
		Set("events_count = ?events_count").
		Set("tx_count = ?tx_count").
		Set("originations_count = ?originations_count").
		Set("sr_originations_count = ?sr_originations_count").
		Set("register_global_constants_count = ?register_global_constants_count").
		Set("sr_executes_count = ?sr_executes_count").
		Set("transfer_tickets_count = ?transfer_tickets_count").
		Set("global_constants_count = ?global_constants_count").
		Set("smart_rollups_count = ?smart_rollups_count").
		Exec(ctx)
	return err
}

func (r Rollback) GetMigrations(ctx context.Context, level int64) (migrations []migration.Migration, err error) {
	err = r.tx.NewSelect().Model(&migrations).
		ColumnExpr("account_id AS contract__account_id, migration.*").
		Where("migration.level = ?", level).
		Join("LEFT JOIN contracts ON contracts.id = contract_id").
		Scan(ctx)
	return
}

func (r Rollback) GetTicketUpdates(ctx context.Context, level int64) (updates []ticket.TicketUpdate, err error) {
	err = r.tx.NewSelect().Model(&updates).
		Where("ticket_update.level = ?", level).
		Relation("Ticket").
		Scan(ctx)
	return
}

func (r Rollback) UpdateTicket(ctx context.Context, ticket ticket.Ticket) error {
	_, err := r.tx.NewUpdate().
		Model(&ticket).
		Where("id = ?id").
		Set("updates_count = updates_count - ?updates_count").
		Exec(ctx)
	return err
}

func (r Rollback) TicketBalances(ctx context.Context, balances ...*ticket.Balance) error {
	if len(balances) == 0 {
		return nil
	}

	_, err := r.tx.NewInsert().Model(&balances).
		Column("ticket_id", "account_id", "amount").
		On("CONFLICT (ticket_id, account_id) DO UPDATE").
		Set("amount = balance.amount - EXCLUDED.amount").
		Exec(ctx)
	return err
}

func (r Rollback) DeleteTickets(ctx context.Context, level int64) (ids []int64, err error) {
	_, err = r.tx.NewDelete().
		Model((*ticket.Ticket)(nil)).
		Where("level = ?", level).
		Returning("id").
		Exec(ctx, &ids)
	return
}

func (r Rollback) DeleteTicketBalances(ctx context.Context, ticketIds []int64) (err error) {
	_, err = r.tx.NewDelete().
		Model((*ticket.Balance)(nil)).
		Where("ticket_id IN (?)", bun.In(ticketIds)).
		Exec(ctx)
	return
}
