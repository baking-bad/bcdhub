package core

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/migration"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/models/stats"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/uptrace/bun"
)

type Transaction struct {
	tx bun.Tx
}

// NewTransaction -
func NewTransaction(ctx context.Context, db *bun.DB) (Transaction, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Transaction{}, err
	}
	return Transaction{tx}, nil
}

func (t Transaction) Commit() error {
	return t.tx.Commit()
}

func (t Transaction) Rollback() error {
	return t.tx.Rollback()
}

func (t Transaction) Save(ctx context.Context, data any) error {
	_, err := t.tx.NewInsert().Model(data).Returning("id").Exec(ctx)
	return err
}

func (t Transaction) Migrations(ctx context.Context, migrations ...*migration.Migration) error {
	if len(migrations) == 0 {
		return nil
	}
	return t.Save(ctx, &migrations)
}

func (t Transaction) GlobalConstants(ctx context.Context, constants ...*contract.GlobalConstant) error {
	if len(constants) == 0 {
		return nil
	}
	return t.Save(ctx, &constants)
}

func (t Transaction) BigMapStates(ctx context.Context, states ...*bigmapdiff.BigMapState) error {
	if len(states) == 0 {
		return nil
	}
	_, err := t.tx.
		NewInsert().
		Model(&states).
		On("CONFLICT ON CONSTRAINT big_map_state_unique DO UPDATE").
		Set("removed = EXCLUDED.removed").
		Set("last_update_level = EXCLUDED.last_update_level").
		Set("last_update_time = EXCLUDED.last_update_time").
		Set("count = big_map_state.count + 1").
		Set("value = CASE WHEN EXCLUDED.removed THEN big_map_state.value ELSE EXCLUDED.value END").
		Returning("id").
		Exec(ctx)
	return err
}

func (t Transaction) BigMapDiffs(ctx context.Context, bigmapdiffs ...*bigmapdiff.BigMapDiff) error {
	if len(bigmapdiffs) == 0 {
		return nil
	}
	return t.Save(ctx, &bigmapdiffs)
}

func (t Transaction) BigMapActions(ctx context.Context, bigmapactions ...*bigmapaction.BigMapAction) error {
	if len(bigmapactions) == 0 {
		return nil
	}
	return t.Save(ctx, &bigmapactions)
}

func (t Transaction) Accounts(ctx context.Context, accounts ...*account.Account) error {
	if len(accounts) == 0 {
		return nil
	}
	_, err := t.tx.NewInsert().Model(&accounts).
		Column("address", "level", "type", "operations_count", "last_action", "events_count", "migrations_count", "ticket_updates_count").
		On("CONFLICT ON CONSTRAINT address_hash DO UPDATE").
		Set("operations_count = EXCLUDED.operations_count + account.operations_count").
		Set("events_count = EXCLUDED.events_count + account.events_count").
		Set("migrations_count = EXCLUDED.migrations_count + account.migrations_count").
		Set("ticket_updates_count = EXCLUDED.ticket_updates_count + account.ticket_updates_count").
		Set("last_action = EXCLUDED.last_action").
		Returning("id").
		Exec(ctx)
	return err
}

func (t Transaction) SmartRollups(ctx context.Context, rollups ...*smartrollup.SmartRollup) error {
	if len(rollups) == 0 {
		return nil
	}
	return t.Save(ctx, &rollups)
}

func (t Transaction) Operations(ctx context.Context, operations ...*operation.Operation) error {
	if len(operations) == 0 {
		return nil
	}
	return t.Save(ctx, &operations)
}

func (t Transaction) TickerUpdates(ctx context.Context, updates ...*ticket.TicketUpdate) error {
	if len(updates) == 0 {
		return nil
	}
	return t.Save(ctx, &updates)
}

func (t Transaction) Contracts(ctx context.Context, contracts ...*contract.Contract) error {
	if len(contracts) == 0 {
		return nil
	}
	return t.Save(ctx, &contracts)
}

func (t Transaction) Scripts(ctx context.Context, scripts ...*contract.Script) error {
	if len(scripts) == 0 {
		return nil
	}
	_, err := t.tx.NewInsert().
		Model(&scripts).
		On("CONFLICT (hash) DO UPDATE").
		Set("tags = EXCLUDED.tags").
		Returning("id").
		Exec(ctx)
	return err
}

func (t Transaction) ScriptConstant(ctx context.Context, relations ...*contract.ScriptConstants) error {
	if len(relations) == 0 {
		return nil
	}
	_, err := t.tx.NewInsert().Model(&relations).Exec(ctx)
	return err
}

func (t Transaction) Block(ctx context.Context, block *block.Block) error {
	if block == nil {
		return nil
	}
	return t.Save(ctx, block)
}

func (t Transaction) Protocol(ctx context.Context, proto *protocol.Protocol) error {
	if proto == nil {
		return nil
	}
	_, err := t.tx.NewInsert().
		Model(proto).
		On("CONFLICT ON CONSTRAINT protocol_hash_idx DO UPDATE").
		Set("end_level = ?", proto.EndLevel).
		Returning("id").
		Exec(ctx)
	return err
}

func (t Transaction) BabylonUpdateNonDelegator(ctx context.Context, contract *contract.Contract) error {
	_, err := t.tx.NewUpdate().
		Model(contract).
		Set("babylon_id = ?babylon_id").
		Where("id = ?id").
		Exec(ctx)
	return err
}

func (t Transaction) JakartaVesting(ctx context.Context, contract *contract.Contract) error {
	_, err := t.tx.NewUpdate().
		Model(contract).
		Set("jakarta_id = babylon_id").
		Where("id = ?id").
		Exec(ctx)
	return err
}

func (t Transaction) JakartaUpdateNonDelegator(ctx context.Context, contract *contract.Contract) error {
	_, err := t.tx.NewUpdate().
		Model(contract).
		Set("jakarta_id = ?jakarta_id").
		Where("id = ?id").
		Exec(ctx)
	return err
}

func (t Transaction) ToBabylon(ctx context.Context) error {
	_, err := t.tx.NewUpdate().Model((*contract.Contract)(nil)).
		Set("babylon_id = alpha_id").
		Where("tags & 4 > 0").
		Exec(ctx)
	return err
}

func (t Transaction) ToJakarta(ctx context.Context) error {
	_, err := t.tx.NewUpdate().Model((*contract.Contract)(nil)).
		Set("jakarta_id = babylon_id").
		Where("tags & 4 > 0").
		Exec(ctx)
	return err
}

func (t Transaction) UpdateStats(ctx context.Context, stats stats.Stats) error {
	_, err := t.tx.NewInsert().
		Model(&stats).
		On("CONFLICT (id) DO UPDATE").
		Set("contracts_count = EXCLUDED.contracts_count + stats.contracts_count").
		Set("operations_count = EXCLUDED.operations_count + stats.operations_count").
		Set("events_count = EXCLUDED.events_count + stats.events_count").
		Set("tx_count = EXCLUDED.tx_count + stats.tx_count").
		Set("originations_count = EXCLUDED.originations_count + stats.originations_count").
		Set("sr_originations_count = EXCLUDED.sr_originations_count + stats.sr_originations_count").
		Set("register_global_constants_count = EXCLUDED.register_global_constants_count + stats.register_global_constants_count").
		Set("sr_executes_count = EXCLUDED.sr_executes_count + stats.sr_executes_count").
		Set("transfer_tickets_count = EXCLUDED.transfer_tickets_count + stats.transfer_tickets_count").
		Set("global_constants_count = EXCLUDED.global_constants_count + stats.global_constants_count").
		Set("smart_rollups_count = EXCLUDED.smart_rollups_count + stats.smart_rollups_count").
		Returning("id").
		Exec(ctx)
	return err
}

func (t Transaction) Tickets(ctx context.Context, tickets ...*ticket.Ticket) error {
	if len(tickets) == 0 {
		return nil
	}

	_, err := t.tx.NewInsert().Model(&tickets).
		Column("ticketer_id", "content", "content_type", "updates_count", "level").
		On("CONFLICT ON CONSTRAINT ticket_key DO UPDATE").
		Set("updates_count = ticket.updates_count + EXCLUDED.updates_count").
		Returning("id").
		Exec(ctx)
	return err
}

func (t Transaction) TicketBalances(ctx context.Context, balances ...*ticket.Balance) error {
	if len(balances) == 0 {
		return nil
	}

	_, err := t.tx.NewInsert().Model(&balances).
		Column("ticket_id", "account_id", "amount").
		On("CONFLICT (ticket_id, account_id) DO UPDATE").
		Set("amount = balance.amount + EXCLUDED.amount").
		Exec(ctx)
	return err
}

func (t Transaction) DeleteBigMapStatesByContract(ctx context.Context, contract string) (states []bigmapdiff.BigMapState, err error) {
	_, err = t.tx.NewDelete().
		Model((*bigmapdiff.BigMapState)(nil)).
		Where("contract = ?", contract).
		Returning("*").
		Exec(ctx, &states)
	return
}

func (t Transaction) BabylonUpdateBigMapDiffs(ctx context.Context, contract string, ptr int64) (int, error) {
	res, err := t.tx.NewUpdate().
		Model((*bigmapdiff.BigMapDiff)(nil)).
		Where("contract = ?", contract).
		Set("ptr = ?", ptr).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	count, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
