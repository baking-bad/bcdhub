package account

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, address string) (Account, error)
	Alias(network types.Network, address string) (string, error)
	UpdateAlias(account Account) error
}
