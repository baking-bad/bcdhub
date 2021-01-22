package bcdast

import "errors"

// Errors
var (
	ErrInvalidPrim        = errors.New("invalid prim")
	ErrInvalidArgsCount   = errors.New("invalid args count")
	ErrUnknownPrim        = errors.New("Unknown prim")
	ErrInvalidJSON        = errors.New("Invalid JSON")
	ErrTreesAreDifferent  = errors.New("Trees of type and value are different")
	ErrEmptyPrim          = errors.New("Empty primitive")
	ErrEmptyUnforgingData = errors.New("Unforging data length is 0")
)
