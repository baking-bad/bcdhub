package bigmapdiff

// Repository -
type Repository interface {
	Get(ctx GetContext) ([]BigMapDiff, error)
	GetByAddress(string, string) ([]BigMapDiff, error)
	GetByOperationID(string) ([]*BigMapDiff, error)
	GetByPtr(string, string, int64) ([]BigMapDiff, error)
	GetByPtrAndKeyHash(int64, string, string, int64, int64) ([]BigMapDiff, int64, error)
	GetForAddress(string) ([]BigMapDiff, error)
	GetValuesByKey(string) ([]BigMapDiff, error)
	GetUniqueByOperationID(string) ([]BigMapDiff, error)
	Count(network string, ptr int64) (int64, error)
	CurrentByKey(network, keyHash string, ptr int64) (BigMapDiff, error)
	Previous([]BigMapDiff, int64, string) ([]BigMapDiff, error)
}
