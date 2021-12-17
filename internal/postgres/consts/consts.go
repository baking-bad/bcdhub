package consts

import "errors"

// Errors
var (
	ErrInvalidAddress     = errors.New("invalid address")
	ErrInvalidPointer     = errors.New("invalid pointer")
	ErrInvalidFingerprint = errors.New("invalid contract fingerprint")
)

// default
const (
	DefaultSize = 10
	MaxSize     = 1000
)
