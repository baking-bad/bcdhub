package functions

var coeffs = []float64{9.697780111000132, 3.9581457911819973, 8.970952588179316, -0.8020228376594127, 2.3415994229548027, 5.250596904928394, -2.4763620053725925, -3.6787311803711433, 3.803540681500497, 8.643552926050244, 11.396978121149688, 6.879624879738669}
var intercepts = -43.33814505431056

// LinearSVC -
type LinearSVC struct {
	coefficients []float64
	intercepts   float64
}

// NewLinearSVC -
func NewLinearSVC() LinearSVC {
	return LinearSVC{
		coefficients: coeffs,
		intercepts:   intercepts,
	}
}

// Predict -
func (svc LinearSVC) Predict(features []float64) int {
	var prob float64
	for i := 0; i < len(svc.coefficients); i++ {
		prob = prob + svc.coefficients[i]*features[i]
	}
	if (prob + svc.intercepts) > 0 {
		return 1
	}
	return 0
}
