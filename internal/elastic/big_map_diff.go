package elastic

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/search"
)

type bigMapDiffSearchResponse struct {
	Aggs struct {
		Diffs struct {
			Buckets []search.BigMapDiffResult `json:"buckets"`
		} `json:"diffs"`
	} `json:"aggregations"`
}

// BigMapDiffs -
func (e *Elastic) BigMapDiffs(args search.BigMapDiffSearchArgs) ([]search.BigMapDiffResult, error) {
	filters := make([]Item, 0)
	if args.Contract != "" {
		filters = append(filters, Match("address", args.Contract))
	}
	if args.Network != types.Empty {
		filters = append(filters, Match("network", args.Network.String()))
	}
	if args.Ptr != nil {
		filters = append(filters, Term("ptr", *args.Ptr))
	}

	orders := make([]Item, 0)
	if args.MinLevel != nil {
		orders = append(orders, Item{
			"gte": args.MinLevel,
		})
	}
	if args.MaxLevel != nil {
		orders = append(orders, Item{
			"lt": args.MaxLevel,
		})
	}
	if len(orders) > 0 {
		filters = append(filters, Range("level", orders...))
	}

	if args.Query != "" {
		filters = append(filters, Bool(
			Should(
				Wildcard("key_hash", fmt.Sprintf("%s*", args.Query)),
				Wildcard("key_strings", fmt.Sprintf("%s*", args.Query)),
			),
			MinimumShouldMatch(1),
		))
	}

	if args.Size == 0 || args.Size > MaxQuerySize {
		args.Size = defaultSize
	}

	query := NewQuery().Query(
		Bool(
			Filter(filters...),
		),
	).Zero().Add(
		Aggs(AggItem{
			Name: "diffs",
			Body: TermsAgg("key_hash.keyword", 1000),
		}),
	)

	var response bigMapDiffSearchResponse
	if err := e.query([]string{models.DocBigMapDiff}, query, &response); err != nil {
		return nil, err
	}

	if args.Offset > 0 {
		if args.Offset < int64(len(response.Aggs.Diffs.Buckets)) {
			response.Aggs.Diffs.Buckets = response.Aggs.Diffs.Buckets[args.Offset:]
		} else {
			return nil, nil
		}
	}

	if args.Size > 0 {
		if args.Size < int64(len(response.Aggs.Diffs.Buckets)) {
			response.Aggs.Diffs.Buckets = response.Aggs.Diffs.Buckets[:args.Size]
		}
	}

	return response.Aggs.Diffs.Buckets, nil
}
