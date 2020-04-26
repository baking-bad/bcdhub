package helpers

import (
	"sort"
)

// Merge2ArraysInt64 -
func Merge2ArraysInt64(a, b []int64) []int64 {
	if len(a) == 0 && len(b) == 0 {
		return []int64{}
	}
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}

	result := make([]int64, 0)
	maxLen := len(a)
	if len(b) > len(a) {
		maxLen = len(b)
		a, b = b, a
	}

	sort.Slice(a, func(i, j int) bool { return a[i] < a[j] })
	sort.Slice(b, func(i, j int) bool { return b[i] < b[j] })

	for j, i := 0, 0; i < maxLen; {
		if j >= len(b) {
			result = append(result, a[i])
			i++
			continue
		}
		if i >= len(a) {
			result = append(result, b[j])
			j++
			continue
		}
		if a[i] < b[j] {
			result = append(result, a[i])
			i++
		} else if a[i] > b[j] {
			result = append(result, b[j])
			j++
		} else {
			result = append(result, a[i])
			i++
			j++
		}
	}
	return result
}

// MaxInt -
func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// MaxInt64 -
func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// MinInt -
func MinInt(a, b int) int {
	if a > b {
		return b
	}
	return a
}
