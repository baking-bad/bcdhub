package metrics

import "math"

func round(x float64) float64 {
	return math.Floor(x*1000000) / 1000000
}
