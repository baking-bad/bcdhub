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
		if c.MaxLevel > 0 {
			filter = append(filter, core.Range("level", core.Item{"lte": c.MaxLevel}))
		}
		if c.MinLevel > 0 {
			filter = append(filter, core.Range("level", core.Item{"gt": c.MinLevel}))
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
	).Sort("level", "desc").All()
}
