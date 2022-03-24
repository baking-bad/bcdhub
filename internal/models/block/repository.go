package block

// Repository -
type Repository interface {
	Get(level int64) (Block, error)
	Last() (Block, error)
	GetNetworkAlias(chainID string) (string, error)
}
