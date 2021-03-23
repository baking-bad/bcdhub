package tokenmetadata

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// Get -
func (storage *Storage) Get(ctx []tokenmetadata.GetContext, size, offset int64) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := storage.DB.Table(models.DocTokenMetadata)
	buildGetTokenMetadataContext(storage.DB, query, ctx...)

	query.Limit(core.GetPageSize(size))
	if offset > 0 {
		query.Offset(int(offset))
	}

	err = query.Find(&tokens).Error
	return
}

// GetAll -
func (storage *Storage) GetAll(ctx ...tokenmetadata.GetContext) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := storage.DB.Table(models.DocTokenMetadata)
	buildGetTokenMetadataContext(storage.DB, query, ctx...)
	err = query.Find(&tokens).Error
	return
}

// GetWithExtras -
func (storage *Storage) GetWithExtras() (tokens []tokenmetadata.TokenMetadata, err error) {
	err = storage.DB.Table(models.DocTokenMetadata).
		Where("extras->'tags' is not null").
		Or("extras->'formats' is not null").
		Or("extras->'creators' is not null").
		Find(&tokens).Error
	return
}
