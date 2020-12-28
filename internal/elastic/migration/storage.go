package migration

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
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

// Get -
func (storage *Storage) Get(network, address string) ([]migration.Migration, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Must(
				core.MatchPhrase("network", network),
				core.MatchPhrase("address", address),
			),
		),
	).Sort("level", "desc").All()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocMigrations}, query, &response); err != nil {
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

// Count -
func (storage *Storage) Count(network, address string) (int64, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
			),
			core.Should(
				core.MatchPhrase("source", address),
				core.MatchPhrase("destination", address),
			),
			core.MinimumShouldMatch(1),
		),
	).Add(
		core.Aggs(
			core.AggItem{
				Name: "migrations_count",
				Body: core.Count("indexed_time"),
			},
		),
	).Zero()

	var response getContractMigrationCountResponse
	err := storage.es.Query([]string{models.DocMigrations}, query, &response)
	return response.Agg.MigrationsCount.Value, err
}
