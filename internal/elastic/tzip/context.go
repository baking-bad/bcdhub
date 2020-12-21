package tzip

import (
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

func buildGetTokenMetadataContext(ctx tzip.GetTokenMetadataContext) interface{} {
	filters := make([]core.Item, 0)

	if ctx.Contract != "" {
		filters = append(filters, core.MatchPhrase("address", ctx.Contract))
	}
	if ctx.Network != "" {
		filters = append(filters, core.Match("network", ctx.Network))
	}
	if ctx.Level.IsFilled() {
		filters = append(filters, core.BuildComparator(ctx.Level))
	}
	if ctx.TokenID != -1 {
		filters = append(filters, core.Term(
			"tokens.static.token_id", ctx.TokenID,
		))
	}
	return core.NewQuery().Query(
		core.Bool(
			core.Filter(filters...),
		),
	).All()
}
