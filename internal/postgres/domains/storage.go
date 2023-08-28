package domains

import (
	"errors"

	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/types"
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

// BigMapDiffs -
func (storage *Storage) BigMapDiffs(lastID, size int64) (result []domains.BigMapDiff, err error) {
	var ids []int64
	query := storage.DB.Model((*bigmapdiff.BigMapDiff)(nil)).Column("id").Order("id asc")
	if lastID > 0 {
		query.Where("big_map_diff.id > ?", lastID)
	}
	if err = query.Limit(storage.GetPageSize(size)).Select(&ids); err != nil {
		return
	}

	if len(ids) == 0 {
		return
	}

	err = storage.DB.Model((*domains.BigMapDiff)(nil)).WhereIn("big_map_diff.id IN (?)", ids).
		Relation("Operation").Relation("Protocol").
		Select(&result)
	return
}

// Same -
func (storage *Storage) Same(network string, c contract.Contract, limit, offset int, availiableNetworks ...string) ([]domains.Same, error) {
	if limit < 1 || limit > 10 {
		limit = 10
	}

	if offset < 1 {
		offset = 0
	}

	if len(availiableNetworks) == 0 {
		availiableNetworks = []string{types.Mainnet.String()}
	}

	script := c.CurrentScript()
	if script == nil {
		return nil, errors.New("invalid contract script")
	}

	var union *pg.Query
	for i, value := range availiableNetworks {
		schema := pg.Safe(value)

		query := storage.DB.Model().
			TableExpr("?.contracts", schema).
			ColumnExpr("? as network", value).
			ColumnExpr("contracts.*").
			ColumnExpr("accounts.address as account__address").
			Join("LEFT JOIN ?.accounts on contracts.account_id = accounts.id", schema).
			Join("LEFT JOIN ?.scripts as alpha on alpha.id = contracts.alpha_id", schema).
			Join("LEFT JOIN ?.scripts as babylon on babylon.id = contracts.babylon_id", schema).
			Join("LEFT JOIN ?.scripts as jakarta on jakarta.id = contracts.jakarta_id", schema).
			Where("alpha.hash = ?", script.Hash).
			WhereOr("babylon.hash = ?", script.Hash).
			WhereOr("jakarta.hash = ?", script.Hash)

		if value == network {
			query.Where("contracts.id != ?", c.ID)
		}

		if i == 0 {
			union = query
		} else {
			union = union.UnionAll(query)
		}
	}

	var same []domains.Same
	err := storage.DB.Model().
		TableExpr("(?) as same", union).
		Limit(limit).
		Offset(offset).
		Select(&same)
	return same, err
}

// SameCount -
func (storage *Storage) SameCount(c contract.Contract, availiableNetworks ...string) (int, error) {
	if len(availiableNetworks) == 0 {
		return 0, nil
	}

	script := c.CurrentScript()
	if script == nil {
		return 0, errors.New("invalid contract script")
	}

	var union *pg.Query
	for i, value := range availiableNetworks {
		schema := pg.Safe(value)

		query := storage.DB.Model().
			TableExpr("?.contracts", schema).
			ColumnExpr("count(*) as c").
			Join("LEFT JOIN ?.scripts as alpha on alpha.id = contracts.alpha_id", schema).
			Join("LEFT JOIN ?.scripts as babylon on babylon.id = contracts.babylon_id", schema).
			Join("LEFT JOIN ?.scripts as jakarta on jakarta.id = contracts.jakarta_id", schema).
			Where("alpha.hash = ?", script.Hash).
			WhereOr("babylon.hash = ?", script.Hash).
			WhereOr("jakarta.hash = ?", script.Hash)

		if i == 0 {
			union = query
		} else {
			union = union.UnionAll(query)
		}
	}

	var count int
	if err := storage.DB.Model().ColumnExpr("sum(c)").TableExpr("(?) as same", union).Select(&count); err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return count - 1, nil
}
