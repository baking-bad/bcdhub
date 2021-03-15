package tokenmetadata

import (
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/restream/reindexer"
)

func buildGetTokenMetadataContext(query *reindexer.Query, ctx ...tokenmetadata.GetContext) {
	if ctx[0].Contract != "" {
		query.Match("address", ctx[0].Contract)
	}
	if ctx[0].Network != "" {
		query.Match("network", ctx[0].Network)
	}
	if ctx[0].TokenID != -1 {
		query.WhereInt64("tokens.static.token_id", reindexer.EQ, ctx[0].TokenID)
	}
}
