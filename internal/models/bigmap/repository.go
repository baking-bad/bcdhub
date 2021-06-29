package bigmap

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, ptr int64, contract string) (*BigMap, error)
	GetByContract(network types.Network, contract string) ([]BigMap, error)
}

// ActionRepository -
type ActionRepository interface {
	Get(network types.Network, ptr int64) ([]Action, error)
}

// DiffRepository -
type DiffRepository interface {
	Get(ctx GetContext) ([]Bucket, error)
	GetForOperation(id int64) ([]*Diff, error)
	GetForOperations(ids ...int64) ([]Diff, error)
	GetByPtrAndKeyHash(ptr int64, network types.Network, keyHash string, size int64, offset int64) ([]Diff, int64, error)
	GetValuesByKey(keyHash string) ([]Diff, error)
	Previous([]Diff) ([]Diff, error)
	Last(network types.Network, ptr int64, keyHash string, skipRemoved bool) (Diff, error)
}

// StateRepository -
type StateRepository interface {
	Current(network types.Network, keyHash string, ptr int64) (State, error)
	GetStats(network types.Network, ptr int64) (Stats, error)
	ChangedAfter(network types.Network, level int64) ([]State, error)
	Keys(ctx GetContext) (states []State, err error)
	GetByPtr(network types.Network, contract string, ptr int64) ([]State, error)
	GetForAddress(network types.Network, address string) ([]State, error)
	Count(network types.Network, ptr int64) (int64, error)
}
