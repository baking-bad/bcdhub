package transfer

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"gorm.io/gorm"
)

func buildGetContext(db *gorm.DB, query *gorm.DB, ctx transfer.GetContext, withSize bool) {
	if query == nil {
		return
	}

	if ctx.Network != "" {
		query.Where("network = ?", ctx.Network)
	}
	if ctx.Address != "" {
		query.Where(
			db.Where("from = ?", ctx.Address).Or("to = ?", ctx.Address),
		)
	}
	if ctx.Start > 0 {
		query.Where("timestamp >= ?", ctx.Start)
	}
	if ctx.End > 0 {
		query.Where("timestamp < ?", ctx.End)
	}
	if ctx.LastID != "" {
		query.Where("last_id < ?", ctx.LastID)
	}
	subQuery := core.OrStringArray(db, ctx.Contracts, "contract")
	if subQuery != nil {
		query.Where(subQuery)
	}
	if ctx.TokenID != nil {
		query.Where("token_id = ?", *ctx.TokenID)
	}
	if ctx.Hash != "" {
		query.Where("hash = ?", ctx.Hash)
	}
	if ctx.Counter != nil {
		query.Where("counter = ?", *ctx.Counter)
	}
	if ctx.Nonce != nil {
		query.Where("nonce = ?", *ctx.Nonce)
	}

	if withSize {
		size := core.GetPageSize(ctx.Size)
		query.Limit(int(size))

		if ctx.Offset > 0 {
			query.Offset(int(ctx.Offset))
		}
	}
	if ctx.SortOrder == "asc" || ctx.SortOrder == "desc" {
		query.Order(fmt.Sprintf("timestamp %s", ctx.SortOrder))
	}
}
