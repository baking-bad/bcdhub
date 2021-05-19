package migration

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, address string) ([]Migration, error)
	Count(network types.Network, address string) (int64, error)
}
