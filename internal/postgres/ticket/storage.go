package ticket

import (
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
func (storage *Storage) Get(ticketer string, limit, offset int64) ([]ticket.TicketUpdate, error) {
	query := storage.DB.
		Model((*ticket.TicketUpdate)(nil)).
		Relation("Ticketer").
		Relation("Account").
		Where("ticketer.address = ?", ticketer)

	if offset > 0 {
		query.Offset(int(offset))
	}
	if limit > 0 {
		query.Limit(storage.GetPageSize(limit))
	}

	var response []ticket.TicketUpdate
	err := query.Order("id desc").Select(&response)
	return response, err
}
