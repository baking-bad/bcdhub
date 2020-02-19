package functions

var coeffs = []float64{17.171338811296536, 11.770951973455645, 14.907480983141763, 1.7162723814258438, 3.281856254487437, 7.715542876431999, 4.353079887635193, 1.3620081082486621, 4.46090907300656, 11.333964126321767, 12.362824344458845, 14.146202045679866}
var intercepts = -87.2006816119435

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
