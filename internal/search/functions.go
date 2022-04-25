package search

import (
	"context"
	"regexp"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

var ptrRegEx = regexp.MustCompile(`^ptr:\d+$`)

// IsPtrSearch - check searchString on `ptr:%d` pattern
func IsPtrSearch(searchString string) bool {
	return ptrRegEx.MatchString(searchString)
}

// Save -
func Save[M models.Constraint](ctx context.Context, searcher Searcher, network types.Network, items []M) error {
	if len(items) == 0 {
		return nil
	}
	data := make([]Data, 0)
	for i := range items {
		if item := Prepare(network, items[i]); item != nil {
			data = append(data, item)
		}
	}
	if len(data) == 0 {
		return nil
	}
	return searcher.Save(ctx, data)
}
