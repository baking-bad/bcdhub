package meta

import "fmt"

// RequiredError -
type RequiredError struct {
	Field string
}

// NewRequiredError -
func NewRequiredError(field string) RequiredError {
	return RequiredError{Field: field}
}

// Error -
func (err RequiredError) Error() string {
	return fmt.Sprintf("'%s' is required", err.Field)
}

// ValidationError -
type ValidationError struct {
	Field string
}

// NewValidationError -
func NewValidationError(field string) ValidationError {
	return ValidationError{Field: field}
}

// Error -
func (err ValidationError) Error() string {
	return fmt.Sprintf("Validation failed: '%s'", err.Field)
}
