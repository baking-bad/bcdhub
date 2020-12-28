package bigmapdiff

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
	"github.com/restream/reindexer"
)

func buildGetContext(ctx bigmapdiff.GetContext, query *reindexer.Query) {
	if ctx.Ptr != nil {
		query = query.WhereInt64("ptr", reindexer.EQ, *ctx.Ptr)
	}
	if ctx.Network != "" {
		query = query.Match("network", ctx.Network)
	}

	if ctx.Query != "" {
		query = query.OpenBracket().
			Where("key", reindexer.LIKE, fmt.Sprintf("%%%s%%", ctx.Query)).
			Where("key_hash", reindexer.LIKE, fmt.Sprintf("%%%s%%", ctx.Query)).
			Where("key_strings", reindexer.LIKE, fmt.Sprintf("%%%s%%", ctx.Query)).
			CloseBracket()
	}

	if ctx.Level != nil {
		query = query.WhereInt64("level", reindexer.LE, *ctx.Level)
	}

	if ctx.Size == 0 {
		ctx.Size = consts.DefaultSize
	}

	query = query.Offset(int(ctx.Offset)).Limit(int(ctx.Size)).Sort("indexed_time", true)
}

// core.Aggs(core.AggItem{
// 	Name: "keys",
// 	Body: core.Item{
// 		"terms": core.Item{
// 			"field": "key_hash.keyword",
// 			"size":  ctx.To,
// 			"order": core.Item{
// 				"bucketsSort": "desc",
// 			},
// 		},
// 		"aggs": core.Item{
// 			"top_key":     core.TopHits(1, "indexed_time", "desc"),
// 			"bucketsSort": core.Max("indexed_time"),
// 		},
// 	},
// }),
