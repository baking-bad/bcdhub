package bigmapdiff

// Repository -
type Repository interface {
	Get(ctx GetContext) ([]Bucket, error)
	GetByAddress(string, string) ([]BigMapDiff, error)
	GetForOperation(hash string, counter int64, nonce *int64) ([]*BigMapDiff, error)
	GetUniqueForOperation(hash string, counter int64, nonce *int64) ([]BigMapDiff, error)
	GetByPtr(string, string, int64) ([]BigMapDiff, error)
	GetByPtrAndKeyHash(int64, string, string, int64, int64) ([]BigMapDiff, int64, error)
	GetForAddress(string) ([]BigMapDiff, error)
	GetByIDs(ids ...int64) ([]BigMapDiff, error)
	GetValuesByKey(string) ([]BigMapDiff, error)
	Count(network string, ptr int64) (int64, error)
	Current(network, keyHash string, ptr int64) (BigMapState, error)
	CurrentByContract(network, contract string) ([]BigMapState, error)
	Previous([]BigMapDiff, int64, string) ([]BigMapDiff, error)
	GetStats(network string, ptr int64) (Stats, error)
}
