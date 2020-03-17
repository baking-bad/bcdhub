package cerrors

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

// HasParametersError -
func HasParametersError(err []Error) bool {
	for i := range err {
		if strings.Contains(err[i].ID, consts.BadParameterError) {
			return true
		}
	}
	return false
}

// HasGasExhaustedError -
func HasGasExhaustedError(err []Error) bool {
	for i := range err {
		if strings.Contains(err[i].ID, consts.GasExhaustedError) {
			return true
		}
	}
	return false
}
