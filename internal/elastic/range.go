package elastic

// Range -
type Range struct {
	Comparator string
	Value      int64
}

// NewRange -
func NewRange(cmp string, value int64) Range {
	return Range{
		Comparator: cmp,
		Value:      value,
	}
}

func (rng Range) build() qItem {
	return rangeQ("level", qItem{
		rng.Comparator: rng.Value,
	})
}

func (rng Range) isFilled() bool {
	return rng.Comparator != "" && rng.Value > 0
}

// NewGreaterThanRange -
func NewGreaterThanRange(value int64) Range {
	return NewRange("gt", value)
}

// NewGreaterThanEqRange -
func NewGreaterThanEqRange(value int64) Range {
	return NewRange("gte", value)
}

// NewLessThanRange -
func NewLessThanRange(value int64) Range {
	return NewRange("lt", value)
}

// NewLessThanEqRange -
func NewLessThanEqRange(value int64) Range {
	return NewRange("lte", value)
}
