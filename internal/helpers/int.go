package helpers

// Number -
type Number interface {
	~int64 | ~float64 | ~int | ~float32
}

// Max -
func Max[T Number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

// Min -
func Min[T Number](a, b T) T {
	if a > b {
		return b
	}
	return a
}

// IsInt64PointersEqual -
func IsInt64PointersEqual(a, b *int64) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a != nil && b != nil:
		return *a == *b
	default:
		return false
	}
}
