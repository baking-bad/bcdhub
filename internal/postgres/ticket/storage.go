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

// Get -
func (storage *Storage) Updates(ctx context.Context, ticketer string, limit, offset int64) (response []ticket.TicketUpdate, err error) {
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
		Model(&response).
		Relation("Ticket").
		Relation("Ticket.Ticketer").
		Relation("Account", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Column("address")
		}).
		Where("ticket.ticketer_id = ?", ticketerId)

	if offset > 0 {
		query.Offset(int(offset))
	}
	if limit > 0 {
		query.Limit(storage.GetPageSize(limit))
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
