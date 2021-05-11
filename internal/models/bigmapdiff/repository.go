package bigmapdiff

// Repository -
type Repository interface {
	Get(ctx GetContext) ([]Bucket, error)
	GetByAddress(string, string) ([]BigMapDiff, error)
	GetForOperation(hash string, counter int64, nonce *int64) ([]*BigMapDiff, error)
	GetUniqueForOperations(opg []OPG) ([]BigMapDiff, error)
	GetByPtr(network, contract string, ptr int64) ([]BigMapState, error)
	GetByPtrAndKeyHash(int64, string, string, int64, int64) ([]BigMapDiff, int64, error)
	GetForAddress(network, address string) ([]BigMapState, error)
	GetByIDs(ids ...int64) ([]BigMapDiff, error)
	GetValuesByKey(string) ([]BigMapDiff, error)
	Count(network string, ptr int64) (int64, error)
	Current(network, keyHash string, ptr int64) (BigMapState, error)
	CurrentByContract(network, contract string) ([]BigMapState, error)
	Previous([]BigMapDiff) ([]BigMapDiff, error)
	GetStats(network string, ptr int64) (Stats, error)
	StatesChangedAfter(network string, level int64) ([]BigMapState, error)
	LastDiff(network string, ptr int64, keyHash string, skipRemoved bool) (BigMapDiff, error)
	Keys(ctx GetContext) (states []BigMapState, err error)
}
