package domains

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	TokenBalances(network types.Network, contract, address string, size, offset int64, sort string) (TokenBalanceResponse, error)
}
