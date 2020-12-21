package core

import "github.com/baking-bad/bcdhub/internal/models/tzip"

// BuildComparator -
func BuildComparator(rng tzip.Comparator) Item {
	return Range("level", Item{
		rng.Comparator: rng.Value,
	})
}

// NewGreaterThanRange -
func NewGreaterThanRange(value int64) tzip.Comparator {
	return tzip.NewRange("gt", value)
}

// NewGreaterThanEqRange -
func NewGreaterThanEqRange(value int64) tzip.Comparator {
	return tzip.NewRange("gte", value)
}

// NewLessThanRange -
func NewLessThanRange(value int64) tzip.Comparator {
	return tzip.NewRange("lt", value)
}

// NewLessThanEqRange -
func NewLessThanEqRange(value int64) tzip.Comparator {
	return tzip.NewRange("lte", value)
}
