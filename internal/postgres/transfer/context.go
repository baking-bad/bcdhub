package transfer

import (
	"fmt"
	"strconv"

	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10/orm"
)

func (storage *Storage) buildGetContext(query *orm.Query, ctx transfer.GetContext, withSize bool) {
	if query == nil {
		return
	}

	if ctx.Network != types.Empty {
		query.Where("network = ?", ctx.Network)
	}
	if ctx.AccountID > -1 {
		query.WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q = q.WhereOr(`from_id = ?`, ctx.AccountID).WhereOr(`to = ?`, ctx.AccountID)
			return q, nil
		})
	}

	switch {
	case ctx.Start > 0 && ctx.End > 0:
		query.Where("timestamp between to_timestamp(?) and to_timestamp(?)", ctx.Start, ctx.End)
	case ctx.Start > 0:
		query.Where("timestamp >= to_timestamp(?)", ctx.Start)
	case ctx.End > 0:
		query.Where("timestamp < to_timestamp(?)", ctx.End)
	}

	if ctx.LastID != "" {
		if id, err := strconv.ParseInt(ctx.LastID, 10, 64); err == nil {
			if ctx.SortOrder == "asc" {
				query.Where("transfer.id > ?", id)
			} else {
				query.Where("transfer.id < ?", id)
			}
		}
	}
	subQuery := core.OrStringArray(query, ctx.Contracts, "contract")
	if subQuery != nil {
		query.Where("(?)", subQuery)
	}
	if ctx.TokenID != nil {
		query.Where("token_id = ?", *ctx.TokenID)
	}
	if ctx.OperationID != nil {
		query.Where("operation_id = ?", *ctx.OperationID)
	}

	if withSize {
		query.Limit(storage.GetPageSize(ctx.Size))

		if ctx.Offset > 0 {
			query.Offset(int(ctx.Offset))
		}
	}
	if ctx.SortOrder == "asc" || ctx.SortOrder == "desc" {
		query.Order(fmt.Sprintf("level %s", ctx.SortOrder))
	} else {
		query.Order("level desc")
	}
}
