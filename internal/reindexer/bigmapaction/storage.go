package bigmapaction

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
	"github.com/restream/reindexer"
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
func (storage *Storage) Get(ptr int64, network string) ([]bigmapaction.BigMapAction, error) {
	query := storage.db.Query(models.DocBigMapActions).
		WhereString("network", reindexer.EQ, network).
		OpenBracket().
		WhereInt64("source_ptr", reindexer.EQ, ptr).
		Or().
		WhereInt64("destination_ptr", reindexer.EQ, ptr).
		CloseBracket().
		Sort("indexed_time", true)

	it := query.Exec()
	defer it.Close()

	if it.Error() != nil {
		return nil, it.Error()
	}

	result := make([]bigmapaction.BigMapAction, 0)
	for it.Next() {
		action := it.Object().(*bigmapaction.BigMapAction)
		result = append(result, *action)
	}

	return result, nil
}
