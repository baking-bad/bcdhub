package tokenmetadata

// Repository -
type Repository interface {
	Get(ctx ...GetContext) ([]TokenMetadata, error)
	GetWithExtras() ([]TokenMetadata, error)
}
