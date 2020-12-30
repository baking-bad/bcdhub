package bigmapdiff

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models/bigmapdiff"
)

type getBigMapDiffsWithKeysResponse struct {
	Agg struct {
		Keys struct {
			Buckets []struct {
				DocCount int64 `json:"doc_count"`
				TopKey   struct {
					Hits core.HitsArray `json:"hits"`
				} `json:"top_key"`
			} `json:"buckets"`
		} `json:"keys"`
	} `json:"aggregations"`
}

type getBigMapDiffsCountResponse struct {
	Agg struct {
		Count core.IntValue `json:"count"`
	} `json:"aggregations"`
}

func buildGetContext(ctx *bigmapdiff.GetContext) core.Base {
	filters := make([]core.Item, 0)

	if ctx.Ptr != nil {
		filters = append(filters, core.Term("ptr", *ctx.Ptr))
	}
	if ctx.Network != "" {
		filters = append(filters, core.Match("network", ctx.Network))
	}

	if ctx.Query != "" {
		filters = append(filters, core.QueryString(fmt.Sprintf("*%s*", ctx.Query), []string{"key", "key_hash", "key_strings", "bin_path", "value", "value_strings"}))
	}

	if ctx.Size == 0 {
		ctx.Size = consts.DefaultSize
	}

	if ctx.Level != nil {
		filters = append(filters, core.BuildComparator(core.NewLessThanEqRange(*ctx.Level)))
	}

	ctx.To = ctx.Size + ctx.Offset
	b := core.Bool(
		core.Must(filters...),
	)
	return core.NewQuery().Query(b).Add(
		core.Aggs(core.AggItem{
			Name: "keys",
			Body: core.Item{
				"terms": core.Item{
					"field": "key_hash.keyword",
					"size":  ctx.To,
					"order": core.Item{
						"bucketsSort": "desc",
					},
				},
				"aggs": core.Item{
					"top_key":     core.TopHits(1, "indexed_time", "desc"),
					"bucketsSort": core.Max("indexed_time"),
				},
			},
		}),
	).Sort("indexed_time", "desc").Zero()
}
