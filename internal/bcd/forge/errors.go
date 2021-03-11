package forge

import "errors"

// errors
var (
	ErrTooFewBytes     = errors.New("Too few bytes")
	ErrInvalidKeyword  = errors.New("Invalid prim keyword")
	ErrUnknownTypeCode = errors.New("Unknown type code")
)
