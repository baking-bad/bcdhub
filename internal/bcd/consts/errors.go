package consts

import "errors"

// Errors
var (
	ErrInvalidPrim         = errors.New("invalid prim")
	ErrInvalidArgsCount    = errors.New("invalid args count")
	ErrUnknownPrim         = errors.New("Unknown prim")
	ErrInvalidJSON         = errors.New("Invalid JSON")
	ErrTreesAreDifferent   = errors.New("Trees of type and value are different")
	ErrEmptyPrim           = errors.New("Empty primitive")
	ErrEmptyUnforgingData  = errors.New("Unforging data length is 0")
	ErrInvalidType         = errors.New("Invalid type")
	ErrJSONDataIsAbsent    = errors.New("JSON data is absent")
	ErrValidation          = errors.New("Validation error")
	ErrTypeIsNotComparable = errors.New("Type is not comparable")
	ErrTreeIsNotSettled    = errors.New("Tree is not settled")
	ErrInvalidAddress      = errors.New("Invalid address")
)
