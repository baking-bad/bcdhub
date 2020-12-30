package block

// Repository -
type Repository interface {
	Get(string, int64) (Block, error)
	Last(string) (Block, error)
	LastByNetworks() ([]Block, error)
	GetNetworkAlias(chainID string) (string, error)
}
