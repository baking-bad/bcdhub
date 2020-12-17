package core

// Comparator -
type Comparator struct {
	Comparator string
	Value      int64
}

// NewRange -
func NewRange(cmp string, value int64) Comparator {
	return Comparator{
		Comparator: cmp,
		Value:      value,
	}
}

// Build -
func (rng Comparator) Build() Item {
	return Range("level", Item{
		rng.Comparator: rng.Value,
	})
}

// IsFilled -
func (rng Comparator) IsFilled() bool {
	return rng.Comparator != "" && rng.Value > 0
}

// NewGreaterThanRange -
func NewGreaterThanRange(value int64) Comparator {
	return NewRange("gt", value)
}

// NewGreaterThanEqRange -
func NewGreaterThanEqRange(value int64) Comparator {
	return NewRange("gte", value)
}

// NewLessThanRange -
func NewLessThanRange(value int64) Comparator {
	return NewRange("lt", value)
}

// NewLessThanEqRange -
func NewLessThanEqRange(value int64) Comparator {
	return NewRange("lte", value)
}
