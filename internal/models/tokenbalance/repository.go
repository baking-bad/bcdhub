package tokenbalance

// Repository -
type Repository interface {
	Get(network, contract, address string, tokenID uint64) (TokenBalance, error)
	GetAccountBalances(network, address, contract string, size, offset int64) ([]TokenBalance, int64, error)
	GetHolders(network, contract string, tokenID uint64) ([]TokenBalance, error)
	Batch(network string, addresses []string) (map[string][]TokenBalance, error)
	CountByContract(network, address string) (map[string]int64, error)
	TokenSupply(network, contract string, tokenID uint64) (supply string, err error)
}
