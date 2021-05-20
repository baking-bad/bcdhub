package protocol

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, hash string, level int64) (Protocol, error)
	GetAll() (response []Protocol, err error)
	GetByNetworkWithSort(network types.Network, sortField, order string) (response []Protocol, err error)
}
