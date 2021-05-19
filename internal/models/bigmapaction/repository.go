package bigmapaction

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, ptr int64) ([]BigMapAction, error)
}
