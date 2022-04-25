package contract_metadata

// Repository -
type Repository interface {
	Get(address string) (*ContractMetadata, error)
	GetWithEvents(updatedAt uint64) ([]ContractMetadata, error)
	GetBySlug(slug string) (*ContractMetadata, error)
	GetAliases() ([]ContractMetadata, error)
	Events(address string) (Events, error)
}
