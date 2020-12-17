package tokenbalance

// Repository -
type Repository interface {
	GetAccountBalances(string, string) ([]TokenBalance, error)
	UpdateTokenBalances(updates []*TokenBalance) error
	GetHolders(network, contract string, tokenID int64) ([]TokenBalance, error)
}
