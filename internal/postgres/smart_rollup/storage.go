package smartrollup

import (
	"context"
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/account"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
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
func (storage *Storage) Get(ctx context.Context, address string) (response smartrollup.SmartRollup, err error) {
	var accountID int64
	if err = storage.DB.NewSelect().
		Model((*account.Account)(nil)).
		Column("id").
		Where("address = ?", address).
		Scan(ctx, &accountID); err != nil {
		return
	}

	err = storage.DB.NewSelect().
		Model(&response).
		Where("address_id = ?", accountID).
		Relation("Address").
		Scan(ctx)
	return
}

// List -
func (storage *Storage) List(ctx context.Context, limit, offset int64, sort string) (response []smartrollup.SmartRollup, err error) {
	query := storage.DB.NewSelect().
		Model(&response).
		Limit(storage.GetPageSize(limit))

	if offset > 0 {
		query.Offset(int(offset))
	}
	lowerSort := strings.ToLower(sort)
	if lowerSort != "asc" && lowerSort != "desc" {
		lowerSort = "desc"
	}
	query.OrderExpr("id ?", bun.Safe(lowerSort))

	err = query.Relation("Address").Scan(ctx)
	return
}
