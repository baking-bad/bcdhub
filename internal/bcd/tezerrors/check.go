package tezerrors

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
)

// HasParametersError -
func HasParametersError(err []*Error) bool {
	return HasError(err, consts.BadParameterError)
}

// HasGasExhaustedError -
func HasGasExhaustedError(err []*Error) bool {
	return HasError(err, consts.GasExhaustedError)
}

// HasScriptRejectedError -
func HasScriptRejectedError(err []*Error) bool {
	return HasError(err, consts.ScriptRejectedError)
}

// HasError -
func HasError(err []*Error, errorID string) bool {
	for i := range err {
		if err[i].Is(errorID) {
			return true
		}
	}
	return false
}

// First -
func First(err []*Error, errorID string) *Error {
	for i := range err {
		if err[i].Is(errorID) {
			return err[i]
		}
	}
	return nil
}
