package domains

import (
	"context"
	"database/sql"
	"errors"

	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/domains"
	"github.com/baking-bad/bcdhub/internal/models/types"
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

func (storage *Storage) scriptIDs(schema bun.Safe, hash string) *bun.SelectQuery {
	return storage.DB.NewSelect().
		TableExpr("?.scripts", schema).
		Column("id").
		Where("hash = ?", hash)
}

// Same -
func (storage *Storage) Same(ctx context.Context, network string, c contract.Contract, limit, offset int, availiableNetworks ...string) ([]domains.Same, error) {
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

	var union *bun.SelectQuery
	for i, value := range availiableNetworks {
		schema := bun.Safe(value)

		query := storage.DB.NewSelect().
			TableExpr("?.contracts", schema).
			ColumnExpr("? as network", value).
			ColumnExpr("contracts.*").
			ColumnExpr("accounts.address as account__address, accounts.last_action as account__last_action").
			Join("LEFT JOIN ?.accounts on contracts.account_id = accounts.id", schema).
			WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
				return sq.
					Where("alpha_id IN (?)", storage.scriptIDs(schema, script.Hash)).
					WhereOr("babylon_id IN (?)", storage.scriptIDs(schema, script.Hash)).
					WhereOr("jakarta_id IN (?)", storage.scriptIDs(schema, script.Hash))
			})

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
	err := storage.DB.NewSelect().
		TableExpr("(?) as same", union).
		Limit(limit).
		Offset(offset).
		Scan(ctx, &same)
	return same, err
}

// SameCount -
func (storage *Storage) SameCount(ctx context.Context, c contract.Contract, availiableNetworks ...string) (int, error) {
	if len(availiableNetworks) == 0 {
		return 0, nil
	}

	script := c.CurrentScript()
	if script == nil {
		return 0, errors.New("invalid contract script")
	}

	var union *bun.SelectQuery
	for i, value := range availiableNetworks {
		schema := bun.Safe(value)

		query := storage.DB.NewSelect().
			TableExpr("?.contracts", schema).
			ColumnExpr("count(*) as c").
			Where("alpha_id IN (?)", storage.scriptIDs(schema, script.Hash)).
			WhereOr("babylon_id IN (?)", storage.scriptIDs(schema, script.Hash)).
			WhereOr("jakarta_id IN (?)", storage.scriptIDs(schema, script.Hash))

		if i == 0 {
			union = query
		} else {
			union = union.UnionAll(query)
		}
	}

	var count int
	if err := storage.DB.NewSelect().ColumnExpr("sum(c)").TableExpr("(?) as same", union).Scan(ctx, &count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return count - 1, nil
}
