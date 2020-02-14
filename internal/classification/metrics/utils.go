package metrics

import "math"

func round(x float64, precision int) float64 {
	mult := math.Pow10(precision)
	return math.Floor(x*mult) / mult
}
