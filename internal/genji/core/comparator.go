package core

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// BuildComparator -
func BuildComparator(field string, cmp tzip.Comparator) fmt.Stringer {
	switch cmp.Comparator {
	case "gt":
		return NewGt(field, cmp.Value)
	case "gte":
		return NewGte(field, cmp.Value)
	case "lt":
		return NewLt(field, cmp.Value)
	case "lte":
		return NewLte(field, cmp.Value)
	case "eq":
		return NewEq(field, cmp.Value)
	}
	return nil
}
