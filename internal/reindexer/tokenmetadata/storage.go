package tokenmetadata

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
)

// Storage -
type Storage struct {
	db *core.Reindexer
}

// NewStorage -
func NewStorage(db *core.Reindexer) *Storage {
	return &Storage{db}
}

// Get -
func (storage *Storage) Get(ctx []tokenmetadata.GetContext, size, offset int64) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := storage.db.Query(models.DocTokenMetadata)
	buildGetTokenMetadataContext(query, ctx...)
	err = storage.db.GetAllByQuery(query, &tokens)
	return
}

// Get -
func (storage *Storage) GetAll(ctx ...tokenmetadata.GetContext) (tokens []tokenmetadata.TokenMetadata, err error) {
	query := storage.db.Query(models.DocTokenMetadata)
	buildGetTokenMetadataContext(query, ctx...)
	err = storage.db.GetAllByQuery(query, &tokens)
	return
}
