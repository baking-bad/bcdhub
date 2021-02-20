package storage

import "errors"

// errors
var (
	ErrTreeIsNotSettled        = errors.New("Tree is not settled")
	ErrInvalidPointer          = errors.New("Invalid pointer")
	ErrUnknownTemporaryPointer = errors.New("Unknown temporary pointer")
)
