package block

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, level int64) (Block, error)
	Last(network types.Network) (Block, error)
	GetNetworkAlias(chainID string) (string, error)
}
