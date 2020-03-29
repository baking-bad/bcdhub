package cerrors

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

// HasParametersError -
func HasParametersError(err []IError) bool {
	for i := range err {
		if err[i].Is(consts.BadParameterError) {
			return true
		}
	}
	return false
}

// HasGasExhaustedError -
func HasGasExhaustedError(err []IError) bool {
	for i := range err {
		if err[i].Is(consts.GasExhaustedError) {
			return true
		}
	}
	return false
}
