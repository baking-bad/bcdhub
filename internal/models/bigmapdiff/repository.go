package bigmapdiff

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(ctx GetContext) ([]Bucket, error)
	GetByAddress(network types.Network, address string) ([]BigMapDiff, error)
	GetForOperation(id int64) ([]*BigMapDiff, error)
	GetForOperations(ids ...int64) ([]BigMapDiff, error)
	GetByPtr(network types.Network, contract string, ptr int64) ([]BigMapState, error)
	GetByPtrAndKeyHash(ptr int64, network types.Network, keyHash string, size int64, offset int64) ([]BigMapDiff, int64, error)
	GetForAddress(network types.Network, address string) ([]BigMapState, error)
	GetByIDs(ids ...int64) ([]BigMapDiff, error)
	GetValuesByKey(keyHash string) ([]BigMapDiff, error)
	Count(network types.Network, ptr int64) (int64, error)
	Current(network types.Network, keyHash string, ptr int64) (BigMapState, error)
	CurrentByContract(network types.Network, contract string) ([]BigMapState, error)
	Previous([]BigMapDiff) ([]BigMapDiff, error)
	GetStats(network types.Network, ptr int64) (Stats, error)
	StatesChangedAfter(network types.Network, level int64) ([]BigMapState, error)
	LastDiff(network types.Network, ptr int64, keyHash string, skipRemoved bool) (BigMapDiff, error)
	Keys(ctx GetContext) (states []BigMapState, err error)
}
