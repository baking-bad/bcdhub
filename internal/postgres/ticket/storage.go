package ticket

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
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

// Updates -
func (storage *Storage) Updates(ctx context.Context, req ticket.UpdatesRequest) (response []ticket.TicketUpdate, err error) {
	query := storage.DB.
		NewSelect().
		Model(&response).
		Relation("Ticket").
		Relation("Ticket.Ticketer").
		Relation("Account", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Column("address")
		}).
		Limit(storage.GetPageSize(req.Limit))

	if req.Ticketer != "" {
		var ticketerId uint64
		if err := storage.DB.NewSelect().
			Model((*account.Account)(nil)).
			Column("id").
			Where("address = ?", req.Ticketer).
			Limit(1).
			Scan(ctx, &ticketerId); err != nil {
			return nil, err
		}
		query.Where("ticket.ticketer_id = ?", ticketerId)
	}

	if req.Account != "" {
		var accountId uint64
		if err := storage.DB.NewSelect().
			Model((*account.Account)(nil)).
			Column("id").
			Where("address = ?", req.Account).
			Limit(1).
			Scan(ctx, &accountId); err != nil {
			return nil, err
		}
		query.Where("account_id = ?", accountId)
	}

	if req.TicketId != nil {
		query.Where("ticket_id = ?", *req.TicketId)
	}

	if req.Offset > 0 {
		query.Offset(int(req.Offset))
	}

	err = query.Order("id desc").Scan(ctx)
	return
}

// ForOperation -
func (storage *Storage) UpdatesForOperation(ctx context.Context, operationId int64) (response []ticket.TicketUpdate, err error) {
	err = storage.DB.
		NewSelect().
		Model(&response).
		Relation("Ticket").
		Relation("Ticket.Ticketer").
		Relation("Account").
		Where("operation_id = ?", operationId).
		Scan(ctx)
	return
}

func (storage *Storage) BalancesForAccount(ctx context.Context, accountId int64, req ticket.BalanceRequest) (balances []ticket.Balance, err error) {
	query := storage.DB.
		NewSelect().
		Model(&balances).
		Relation("Ticket").
		Relation("Ticket.Ticketer", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Column("address")
		}).
		Where("account_id = ?", accountId)

	if req.Offset > 0 {
		query.Offset(int(req.Offset))
	}

	if req.Limit > 0 && req.Limit < 100 {
		query.Limit(int(req.Limit))
	} else {
		query.Limit(10)
	}

	if req.WithoutZeroBalances {
		query.Where("amount > 0")
	}

	err = query.Scan(ctx)
	return
}

func (storage *Storage) List(ctx context.Context, ticketer string, limit, offset int64) (tickets []ticket.Ticket, err error) {
	var ticketerId uint64
	if err := storage.DB.NewSelect().
		Model((*account.Account)(nil)).
		Column("id").
		Where("address = ?", ticketer).
		Limit(1).
		Scan(ctx, &ticketerId); err != nil {
		return nil, err
	}

	query := storage.DB.
		NewSelect().
		Model(&tickets).
		Where("ticketer_id = ?", ticketerId).
		Relation("Ticketer")

	if offset > 0 {
		query.Offset(int(offset))
	}

	if limit > 0 && limit < 100 {
		query.Limit(int(limit))
	} else {
		query.Limit(10)
	}

	err = query.Scan(ctx)
	return
}
