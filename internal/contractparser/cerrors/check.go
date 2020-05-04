package cerrors

import (
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
)

// HasParametersError -
func HasParametersError(err []IError) bool {
	return HasError(err, consts.BadParameterError)
}

// HasGasExhaustedError -
func HasGasExhaustedError(err []IError) bool {
	return HasError(err, consts.GasExhaustedError)
}

// HasScriptRejectedError -
func HasScriptRejectedError(err []IError) bool {
	return HasError(err, consts.ScriptRejectedError)
}

// HasError -
func HasError(err []IError, errorID string) bool {
	for i := range err {
		if err[i].Is(errorID) {
			return true
		}
	}
	return false
}

// First -
func First(err []IError, errorID string) IError {
	for i := range err {
		if err[i].Is(errorID) {
			return err[i]
		}
	}
	return nil
}
