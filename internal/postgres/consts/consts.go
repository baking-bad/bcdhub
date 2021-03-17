package consts

import "errors"

// Errors
var (
	ErrInvalidAddress     = errors.New("Invalid address")
	ErrInvalidPointer     = errors.New("Invalid pointer")
	ErrInvalidFingerprint = errors.New("Invalid contract fingerprint")
)

// default
const (
	DefaultSize = 10
	MaxSize     = 100
)
