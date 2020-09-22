package tokens

import "github.com/pkg/errors"

// Exported errors
var (
	ErrNoTokenMetadataRegistryMethod = errors.Errorf("No token_metadata_registry entrypoint")
	ErrNoViewAddressContract         = errors.Errorf("No view address contract")
	ErrInvalidContractParameter      = errors.Errorf("Invalid operation parameter in simulation")
	ErrInvalidRegistryAddress        = errors.Errorf("Invalid registry address")
	ErrNoMetadataKeyInStorage        = errors.Errorf("No token_metadata key in registry storage")
	ErrUnknownNetwork                = errors.Errorf("Unknown network")
	ErrInvalidStorageStructure       = errors.Errorf("Invalid storage structure")
)
