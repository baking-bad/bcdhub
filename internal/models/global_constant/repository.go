package global_constant

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, address string) (GlobalConstant, error)
	All(network types.Network, addresses ...string) ([]GlobalConstant, error)
}
