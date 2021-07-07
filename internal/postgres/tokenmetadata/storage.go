package tokenmetadata

import (
	"errors"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"gorm.io/gorm"
)

// Storage -
type Storage struct {
	*core.Postgres
}

// NewStorage -
func NewStorage(pg *core.Postgres) *Storage {
	return &Storage{pg}
}

// GetOne -
func (storage *Storage) GetOne(network types.Network, contract string, tokenID uint64) (*tokenmetadata.TokenMetadata, error) {
	var metadata tokenmetadata.TokenMetadata
	if err := storage.DB.Model(&tokenmetadata.TokenMetadata{}).Scopes(core.Token(network, contract, tokenID)).First(&metadata).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &metadata, nil
}

// Get -
func (storage *Storage) Get(ctx []tokenmetadata.GetContext, size, offset int64) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := storage.DB.Table(models.DocTokenMetadata)
	storage.buildGetTokenMetadataContext(query, ctx...)

	query.Limit(storage.GetPageSize(size))
	if offset > 0 {
		query.Offset(int(offset))
	}

	err = query.Order("id desc").Find(&tokens).Error
	return
}

// GetAll -
func (storage *Storage) GetAll(ctx ...tokenmetadata.GetContext) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := storage.DB.Table(models.DocTokenMetadata)
	storage.buildGetTokenMetadataContext(query, ctx...)
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

// Count -
func (storage *Storage) Count(ctx []tokenmetadata.GetContext) (count int64, err error) {
	query := storage.DB.Table(models.DocTokenMetadata)
	storage.buildGetTokenMetadataContext(query, ctx...)
	err = query.Count(&count).Error
	return
}
