package smartrollup

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/models/account"
	smartrollup "github.com/baking-bad/bcdhub/internal/models/smart_rollup"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10"
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
func (storage *Storage) Get(address string) (response smartrollup.SmartRollup, err error) {
	var accountID int64
	if err = storage.DB.Model((*account.Account)(nil)).Column("id").Where("address = ?", address).Select(&accountID); err != nil {
		return
	}

	err = storage.DB.Model(&response).Where("address_id = ?", accountID).Relation("Address").Select()
	return
}

// List -
func (storage *Storage) List(limit, offset int64, sort string) (response []smartrollup.SmartRollup, err error) {
	query := storage.DB.Model((*smartrollup.SmartRollup)(nil)).
		Limit(storage.GetPageSize(limit))

	if offset > 0 {
		query.Offset(int(offset))
	}
	lowerSort := strings.ToLower(sort)
	if lowerSort != "asc" && lowerSort != "desc" {
		lowerSort = "desc"
	}
	query.OrderExpr("id ?", pg.Safe(lowerSort))

	err = query.Relation("Address").Select(&response)
	return
}
