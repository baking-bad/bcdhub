package bigmapdiff

// Repository -
type Repository interface {
	GetBigMapKey(network, keyHash string, ptr int64) (BigMapDiff, error)
	// GetBigMapKeys(ctx GetBigMapKeysContext) ([]BigMapDiff, error)
	GetBigMapsForAddress(string, string) ([]BigMapDiff, error)
	GetBigMapValuesByKey(string) ([]BigMapDiff, error)
	GetBigMapDiffsCount(network string, ptr int64) (int64, error)
	GetBigMapDiffsForAddress(string) ([]BigMapDiff, error)
	GetBigMapDiffsPrevious([]BigMapDiff, int64, string) ([]BigMapDiff, error)
	GetBigMapDiffsUniqueByOperationID(string) ([]BigMapDiff, error)
	GetBigMapDiffsByPtrAndKeyHash(int64, string, string, int64, int64) ([]BigMapDiff, int64, error)
	GetBigMapDiffsByOperationID(string) ([]*BigMapDiff, error)
	GetBigMapDiffsByPtr(string, string, int64) ([]BigMapDiff, error)
}
