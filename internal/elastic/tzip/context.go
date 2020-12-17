package tzip

import "github.com/baking-bad/bcdhub/internal/elastic/core"

// GetTokenMetadataContext -
type GetTokenMetadataContext struct {
	Contract string
	Network  string
	TokenID  int64
	Level    core.Comparator
}

// Build -
func (ctx GetTokenMetadataContext) Build() interface{} {
	filters := make([]core.Item, 0)

	if ctx.Contract != "" {
		filters = append(filters, core.MatchPhrase("address", ctx.Contract))
	}
	if ctx.Network != "" {
		filters = append(filters, core.Match("network", ctx.Network))
	}
	if ctx.Level.IsFilled() {
		filters = append(filters, ctx.Level.Build())
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
