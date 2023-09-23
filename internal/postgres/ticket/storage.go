package ticket

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/ticket"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
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
func (storage *Storage) Get(ctx context.Context, ticketer string, limit, offset int64) (response []ticket.TicketUpdate, err error) {
	query := storage.DB.
		NewSelect().
		Model(&response).
		Relation("Ticketer").
		Relation("Account").
		Where("ticketer.address = ?", ticketer)

	if offset > 0 {
		query.Offset(int(offset))
	}
	if limit > 0 {
		query.Limit(storage.GetPageSize(limit))
	}

	err = query.Order("id desc").Scan(ctx)
	return
}

// Has -
func (storage *Storage) Has(ctx context.Context, contractID int64) (bool, error) {
	var id int64
	err := storage.DB.
		NewSelect().
		Model((*ticket.TicketUpdate)(nil)).
		Column("id").
		Where("ticketer_id = ?", contractID).
		OrderExpr("id ASC").
		Limit(1).
		Scan(ctx, &id)

	if err != nil {
		if storage.IsRecordNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ForOperation -
func (storage *Storage) ForOperation(ctx context.Context, operationId int64) (response []ticket.TicketUpdate, err error) {
	err = storage.DB.
		NewSelect().
		Model(&response).
		Relation("Ticketer").
		Relation("Account").
		Where("operation_id = ?", operationId).
		Scan(ctx)
	return
}
