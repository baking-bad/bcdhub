package core

import (
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/restream/reindexer"
)

// SetComaparator -
func SetComaparator(field string, cmp tzip.Comparator, query *reindexer.Query) {
	switch cmp.Comparator {
	case "gt":
		query = query.WhereInt64(field, reindexer.GT, cmp.Value)
	case "gte":
		query = query.WhereInt64(field, reindexer.GE, cmp.Value)
	case "lt":
		query = query.WhereInt64(field, reindexer.LT, cmp.Value)
	case "lte":
		query = query.WhereInt64(field, reindexer.LE, cmp.Value)
	case "eq":
		query = query.WhereInt64(field, reindexer.EQ, cmp.Value)
	}
}
