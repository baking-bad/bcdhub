package elastic

import "github.com/baking-bad/bcdhub/internal/models"

// GetMigrations -
func (e *Elastic) GetMigrations(network, address string) ([]models.Migration, error) {
	query := newQuery().Query(
		boolQ(
			must(
				matchPhrase("network", network),
				matchPhrase("address", address),
			),
		),
	).Sort("level", "desc").All()

	data, err := e.query([]string{DocMigrations}, query)
	if err != nil {
		return nil, err
	}

	migrations := make([]models.Migration, 0)
	for _, hit := range data.Get("hits.hits").Array() {
		var migration models.Migration
		migration.ParseElasticJSON(hit)
		migrations = append(migrations, migration)
	}
	return migrations, nil
}
