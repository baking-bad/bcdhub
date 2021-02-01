package tokenmetadata

import (
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
)

func buildGetTokenMetadataContext(ctx ...tokenmetadata.GetContext) core.Base {
	filters := make([]core.Item, 0)

	for _, c := range ctx {
		filter := make([]core.Item, 0)
		if c.Contract != "" {
			filter = append(filter, core.MatchPhrase("contract", c.Contract))
		}
		if c.Network != "" {
			filter = append(filter, core.Match("network", c.Network))
		}
		if c.Level.IsFilled() {
			filter = append(filter, core.BuildComparator(c.Level))
		}
		if c.TokenID != -1 {
			filter = append(filter, core.Term("token_id", c.TokenID))
		}

		filters = append(filters, core.Bool(
			core.Filter(filter...),
		))
	}

	return core.NewQuery().Query(
		core.Bool(
			core.Should(filters...),
			core.MinimumShouldMatch(1),
		),
	).All()
}
