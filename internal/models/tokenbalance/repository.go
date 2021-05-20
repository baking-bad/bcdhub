package tokenbalance

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, contract, address string, tokenID uint64) (TokenBalance, error)
	GetHolders(network types.Network, contract string, tokenID uint64) ([]TokenBalance, error)
	Batch(network types.Network, addresses []string) (map[string][]TokenBalance, error)
	CountByContract(network types.Network, address string) (map[string]int64, error)
	TokenSupply(network types.Network, contract string, tokenID uint64) (supply string, err error)
}
