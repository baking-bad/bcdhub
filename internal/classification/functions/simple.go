package functions

// Simple -
type Simple struct {
	coeffs     []float64
	intercepts float64
}

// NewSimple -
func NewSimple() Simple {
	return Simple{
		coeffs: []float64{
			0.15, 0.1, 0.1, 0.1, 0.05, 0.05, 0.05, 0.05, 0.05, 0.1, 0.1, 0.1,
		},
		intercepts: 0.85,
	}
}

// Predict -
func (s Simple) Predict(features []float64) int {
	var prob float64
	for i := 0; i < len(s.coeffs); i++ {
		prob = prob + s.coeffs[i]*features[i]
	}
	if (prob + s.intercepts) > 0 {
		return 1
	}
	return 0
}
