package tokenbalance

// Repository -
type Repository interface {
	GetAccountBalances(string, string) ([]TokenBalance, error)
	Update(updates []*TokenBalance) error
	GetHolders(network, contract string, tokenID int64) ([]TokenBalance, error)
	BurnNft(network, contract string, tokenID int64) error
}
