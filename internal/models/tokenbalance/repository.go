package tokenbalance

// Repository -
type Repository interface {
	Get(contract string, accountID int64, tokenID uint64) (TokenBalance, error)
	GetHolders(contract string, tokenID uint64) ([]TokenBalance, error)
	Batch(accountIDs []int64) (map[string][]TokenBalance, error)
	CountByContract(accountID int64, hideEmpty bool) (map[string]int64, error)
	TokenSupply(contract string, tokenID uint64) (supply string, err error)
}
