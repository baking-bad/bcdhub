package tokenbalance

// Repository -
type Repository interface {
	GetAccountBalances(network string, address string, size, offset int64) ([]TokenBalance, int64, error)
	Update(updates []*TokenBalance) error
	GetHolders(network, contract string, tokenID int64) ([]TokenBalance, error)
	BurnNft(network, contract string, tokenID int64) error
}
