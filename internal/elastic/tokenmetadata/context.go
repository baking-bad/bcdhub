package tokenmetadata

import (
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
)

func buildGetTokenMetadataContext(ctx tokenmetadata.GetContext) core.Base {
	filters := make([]core.Item, 0)

	if ctx.Contract != "" {
		filters = append(filters, core.MatchPhrase("contract", ctx.Contract))
	}
	if ctx.Network != "" {
		filters = append(filters, core.Match("network", ctx.Network))
	}
	if ctx.Level.IsFilled() {
		filters = append(filters, core.BuildComparator(ctx.Level))
	}
	if ctx.TokenID != -1 {
		filters = append(filters, core.Term("token_id", ctx.TokenID))
	}
	return core.NewQuery().Query(
		core.Bool(
			core.Filter(filters...),
		),
	).All()
}
