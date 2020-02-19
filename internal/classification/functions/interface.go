package functions

// Predictable -
type Predictable interface {
	Predict(features []float64) int
}
