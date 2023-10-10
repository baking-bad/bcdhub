package account

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/account"
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
func (storage *Storage) Get(ctx context.Context, address string) (account account.Account, err error) {
	err = storage.DB.NewSelect().Model(&account).
		Where("address = ?", address).
		Limit(1).
		Scan(ctx)
	return
}

// RecentlyCalled -
func (storage *Storage) RecentlyCalledContracts(ctx context.Context, offset, size int64) (accounts []account.Account, err error) {
	query := storage.DB.NewSelect().Model(&accounts).
		Where("type = 1")

	if offset > 0 {
		query.Offset(int(offset))
	}
	if size > 0 {
		query.Limit(int(size))
	} else {
		query.Limit(10)
	}
	err = query.
		OrderExpr("last_action desc, operations_count desc").
		Scan(ctx)
	return
}
