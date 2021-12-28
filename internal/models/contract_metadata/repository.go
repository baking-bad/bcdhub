package contract_metadata

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, address string) (*ContractMetadata, error)
	GetWithEvents(updatedAt uint64) ([]ContractMetadata, error)
	GetBySlug(slug string) (*ContractMetadata, error)
	GetAliases(network types.Network) ([]ContractMetadata, error)
	Events(network types.Network, address string) (Events, error)
}
