package protocol

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/protocol"
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

// Get - returns current protocol for `level` (`hash` is optional, leave empty string for default)
func (storage *Storage) Get(ctx context.Context, hash string, level int64) (p protocol.Protocol, err error) {
	query := storage.DB.NewSelect().Model(&p)
	if level > -1 {
		query = query.Where("start_level <= ?", level)
	}
	if hash != "" {
		query = query.Where("hash = ?", hash)
	}

	err = query.Order("start_level DESC").Limit(1).Scan(ctx)
	return
}

// GetByID - returns protocol by id
func (storage *Storage) GetByID(ctx context.Context, id int64) (response protocol.Protocol, err error) {
	err = storage.DB.NewSelect().Model(&response).Where("id = ?", id).Limit(1).Scan(ctx)
	return
}
