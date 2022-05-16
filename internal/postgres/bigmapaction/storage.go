package bigmapaction

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/postgres/core"
	"github.com/go-pg/pg/v10/orm"
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
func (storage *Storage) Get(ptr, limit, offset int64) (actions []bigmapaction.BigMapAction, err error) {
	query := storage.DB.Model().Table(models.DocBigMapActions).
		WhereGroup(func(q *orm.Query) (*orm.Query, error) {
			q.Where("source_ptr = ?", ptr).WhereOr("destination_ptr = ?", ptr)
			return q, nil
		}).
		Order("id DESC")

	if limit > 0 {
		query.Limit(int(limit))
	}
	if offset > 0 {
		query.Offset(int(offset))
	}
	err = query.Select(&actions)
	return
}
