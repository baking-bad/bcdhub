package tokenmetadata

import (
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
	"github.com/restream/reindexer"
)

func buildGetTokenMetadataContext(ctx tokenmetadata.GetContext, query *reindexer.Query) {
	if ctx.Contract != "" {
		query.Match("address", ctx.Contract)
	}
	if ctx.Network != "" {
		query.Match("network", ctx.Network)
	}
	if ctx.Level.IsFilled() {
		core.SetComaparator("level", ctx.Level, query)
	}
	if ctx.TokenID != -1 {
		query.WhereInt64("tokens.static.token_id", reindexer.EQ, ctx.TokenID)
	}
}
