package normalize

import "errors"

// errors
var (
	ErrDataIsNil          = errors.New("Data is nil")
	ErrInvalidJSON        = errors.New("Invalid JSON")
	ErrInvalidDataType    = errors.New("Invalid data type")
	ErrArgsAreAbsent      = errors.New("Args are absent")
	ErrInvalidPrimitive   = errors.New("Invalid primitive")
	ErrInvalidArrayLength = errors.New("Invalid array length")
)
