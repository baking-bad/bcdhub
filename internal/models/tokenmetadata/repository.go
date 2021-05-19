package tokenmetadata

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(ctx []GetContext, size, offset int64) ([]TokenMetadata, error)
	GetAll(ctx ...GetContext) ([]TokenMetadata, error)
	GetOne(network types.Network, contract string, tokenID uint64) (*TokenMetadata, error)
	GetWithExtras() ([]TokenMetadata, error)
	Count(ctx []GetContext) (int64, error)
}
