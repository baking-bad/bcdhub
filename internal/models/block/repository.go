package block

// Repository -
type Repository interface {
	GetBlock(string, int64) (Block, error)
	GetLastBlock(string) (Block, error)
	GetLastBlocks() ([]Block, error)
	GetNetworkAlias(chainID string) (string, error)
}
