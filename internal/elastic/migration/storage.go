package migration

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models/migration"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

// GetMigrations -
func (storage *Storage) GetMigrations(network, address string) ([]migration.Migration, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Must(
				core.MatchPhrase("network", network),
				core.MatchPhrase("address", address),
			),
		),
	).Sort("level", "desc").All()

	var response core.SearchResponse
	if err := storage.es.Query([]string{consts.DocMigrations}, query, &response); err != nil {
		return nil, err
	}

	migrations := make([]migration.Migration, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &migrations[i]); err != nil {
			return nil, err
		}
	}
	return migrations, nil
}
