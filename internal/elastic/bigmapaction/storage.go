package bigmapaction

import (
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models/bigmapaction"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

// Get -
func (storage *Storage) Get(ptr int64, network string) (response []bigmapaction.BigMapAction, err error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
			),
			core.Should(
				core.Term("source_ptr", ptr),
				core.Term("destination_ptr", ptr),
			),
			core.MinimumShouldMatch(1),
		),
	).Sort("indexed_time", "desc")

	err = storage.es.GetAllByQuery(query, &response)
	return
}
