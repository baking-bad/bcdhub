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
