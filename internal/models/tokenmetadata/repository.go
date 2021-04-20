package tokenmetadata

// Repository -
type Repository interface {
	Get(ctx []GetContext, size, offset int64) ([]TokenMetadata, error)
	GetAll(ctx ...GetContext) ([]TokenMetadata, error)
	GetOne(network, contract string, tokenID int64) (token TokenMetadata, err error)
	GetWithExtras() ([]TokenMetadata, error)
	Count(ctx []GetContext) (int64, error)
}
