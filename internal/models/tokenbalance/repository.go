package tokenbalance

// Repository -
type Repository interface {
	GetAccountBalances(network, address, contract string, size, offset int64) ([]TokenBalance, int64, error)
	Update(updates []*TokenBalance) error
	GetHolders(network, contract string, tokenID int64) ([]TokenBalance, error)
	Batch(network string, addresses []string) (map[string][]TokenBalance, error)
	BurnNft(network, contract string, tokenID int64) error
	CountByContract(network, address string) (map[string]int64, error)
}
