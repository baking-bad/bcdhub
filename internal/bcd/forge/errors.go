package forge

import "errors"

// errors
var (
	ErrTooFewBytes     = errors.New("too few bytes")
	ErrInvalidKeyword  = errors.New("invalid prim keyword")
	ErrUnknownTypeCode = errors.New("unknown type code")
)
