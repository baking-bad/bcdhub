package core

import (
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
)

// BuildComparator -
func BuildComparator(rng tokenmetadata.Comparator) Item {
	return Range("level", Item{
		rng.Comparator: rng.Value,
	})
}

// NewGreaterThanRange -
func NewGreaterThanRange(value int64) tokenmetadata.Comparator {
	return tokenmetadata.NewRange("gt", value)
}

// NewGreaterThanEqRange -
func NewGreaterThanEqRange(value int64) tokenmetadata.Comparator {
	return tokenmetadata.NewRange("gte", value)
}

// NewLessThanRange -
func NewLessThanRange(value int64) tokenmetadata.Comparator {
	return tokenmetadata.NewRange("lt", value)
}

// NewLessThanEqRange -
func NewLessThanEqRange(value int64) tokenmetadata.Comparator {
	return tokenmetadata.NewRange("lte", value)
}
