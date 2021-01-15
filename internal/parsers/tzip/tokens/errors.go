package tokens

import "github.com/pkg/errors"

// Exported errors
var (
	ErrNoMetadataKeyInStorage  = errors.Errorf("No token_metadata key in storage")
	ErrInvalidStorageStructure = errors.Errorf("Invalid storage structure")
)
