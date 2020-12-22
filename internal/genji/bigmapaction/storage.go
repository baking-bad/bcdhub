package bigmapaction

import (
	"github.com/baking-bad/bcdhub/internal/genji/core"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
)

// Storage -
type Storage struct {
	db *core.Genji
}

// NewStorage -
func NewStorage(db *core.Genji) *Storage {
	return &Storage{db}
}

// Get -
func (storage *Storage) Get(ptr int64, network string) (response []bigmapaction.BigMapAction, err error) {
	builder := core.NewBuilder().And(
		core.NewEq("network", network),
		core.NewOr(
			core.NewEq("source_ptr", ptr),
			core.NewEq("destination_ptr", ptr),
		),
	).SortDesc("indexed_time")

	err = storage.db.GetAllByQuery(builder, &response)
	return
}
