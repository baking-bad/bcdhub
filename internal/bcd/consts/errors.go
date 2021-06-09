package consts

import "errors"

// Errors
var (
	ErrInvalidPrim         = errors.New("invalid prim")
	ErrInvalidArgsCount    = errors.New("invalid args count")
	ErrUnknownPrim         = errors.New("unknown prim")
	ErrInvalidJSON         = errors.New("invalid JSON")
	ErrTreesAreDifferent   = errors.New("trees of type and value are different")
	ErrEmptyPrim           = errors.New("empty primitive")
	ErrEmptyUnforgingData  = errors.New("unforging data length is 0")
	ErrInvalidType         = errors.New("invalid type")
	ErrJSONDataIsAbsent    = errors.New("JSON data is absent")
	ErrValidation          = errors.New("validation error")
	ErrTypeIsNotComparable = errors.New("type is not comparable")
	ErrTreeIsNotSettled    = errors.New("tree is not settled")
	ErrInvalidAddress      = errors.New("invalid address")
	ErrInvalidOPGHash      = errors.New("invalid OPG hash")
	ErrUnknownPointer      = errors.New("can't find ptr")
	ErrEmptyTree           = errors.New("empty tree")
)
