package bigmapdiff

// Repository -
type Repository interface {
	Get(ctx GetContext) ([]Bucket, error)
	GetByAddress(address string) ([]BigMapDiff, error)
	GetForOperation(id int64) ([]BigMapDiff, error)
	GetByPtr(contract string, ptr int64) ([]BigMapState, error)
	GetByPtrAndKeyHash(ptr int64, keyHash string, size int64, offset int64) ([]BigMapDiff, int64, error)
	GetForAddress(address string) ([]BigMapState, error)
	GetValuesByKey(keyHash string) ([]BigMapState, error)
	Count(ptr int64) (int64, error)
	Current(keyHash string, ptr int64) (BigMapState, error)
	CurrentByContract(contract string) ([]BigMapState, error)
	Previous([]BigMapDiff) ([]BigMapDiff, error)
	GetStats(ptr int64) (Stats, error)
	StatesChangedAfter(level int64) ([]BigMapState, error)
	LastDiff(ptr int64, keyHash string, skipRemoved bool) (BigMapDiff, error)
	Keys(ctx GetContext) (states []BigMapState, err error)
}
