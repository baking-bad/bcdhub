package storage

import "errors"

// errors
var (
	ErrTreeIsNotSettled        = errors.New("tree is not settled")
	ErrInvalidPointer          = errors.New("invalid pointer")
	ErrUnknownTemporaryPointer = errors.New("unknown temporary pointer")
)
