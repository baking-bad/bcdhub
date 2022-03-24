package tokenmetadata

import (
	"errors"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10"
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
func (storage *Storage) GetOne(contract string, tokenID uint64) (*tokenmetadata.TokenMetadata, error) {
	var metadata tokenmetadata.TokenMetadata
	query := storage.DB.Model(&tokenmetadata.TokenMetadata{})
	core.Token(contract, tokenID)(query)

	if err := query.First(); err != nil {
		if errors.Is(err, pg.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &metadata, nil
}

// Get -
func (storage *Storage) Get(ctx []tokenmetadata.GetContext, size, offset int64) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := storage.DB.Model(&tokenmetadata.TokenMetadata{})
	storage.buildGetTokenMetadataContext(query, ctx...)

	query.Limit(storage.GetPageSize(size))
	if offset > 0 {
		query.Offset(int(offset))
	}

	err = query.Order("id desc").Select(&tokens)
	return
}

// GetAll -
func (storage *Storage) GetAll(ctx ...tokenmetadata.GetContext) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := storage.DB.Model(&tokenmetadata.TokenMetadata{})
	storage.buildGetTokenMetadataContext(query, ctx...)
	err = query.Select(&tokens)
	return
}

// GetRecent -
func (storage *Storage) GetRecent(since time.Time, ctx ...tokenmetadata.GetContext) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := storage.DB.Model(&tokenmetadata.TokenMetadata{})
	storage.buildGetTokenMetadataContext(query, ctx...)
	err = query.
		Where("timestamp > ?", since).
		Order("id desc").
		Select(&tokens)
	return
}

// GetWithExtras -
func (storage *Storage) GetWithExtras() (tokens []tokenmetadata.TokenMetadata, err error) {
	err = storage.DB.Model(&tokenmetadata.TokenMetadata{}).
		Where("extras->'tags' is not null").
		WhereOr("extras->'formats' is not null").
		WhereOr("extras->'creators' is not null").
		Select(&tokens)
	return
}

// Count -
func (storage *Storage) Count(ctx []tokenmetadata.GetContext) (int64, error) {
	query := storage.DB.Model(&tokenmetadata.TokenMetadata{})
	storage.buildGetTokenMetadataContext(query, ctx...)
	count, err := query.Count()
	return int64(count), err
}
