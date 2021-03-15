package tokenmetadata

// Repository -
type Repository interface {
	Get(ctx []GetContext, size, offset int64) ([]TokenMetadata, error)
	GetAll(ctx ...GetContext) ([]TokenMetadata, error)
}
