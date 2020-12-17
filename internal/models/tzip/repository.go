package tzip

// Repository -
type Repository interface {
	Get(network, address string) (TZIP, error)
	GetWithEvents() ([]TZIP, error)
	GetTokenMetadata(ctx GetTokenMetadataContext) ([]TokenMetadata, error)
	GetDApps() ([]DApp, error)
	GetDAppBySlug(slug string) (*DApp, error)
	GetBySlug(slug string) (*TZIP, error)
	GetAliases(network string) ([]TZIP, error)
	GetAliasesMap(network string) (map[string]string, error)
	GetAlias(network, address string) (*TZIP, error)
}
