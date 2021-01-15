package tokenmetadata

// GetContext -
type GetContext struct {
	Contract string
	Network  string
	TokenID  int64
	Level    Comparator
}

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

// IsFilled -
func (rng Comparator) IsFilled() bool {
	return rng.Comparator != "" && rng.Value > 0
}
