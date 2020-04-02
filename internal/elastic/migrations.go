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

// GetContractVersions -
func (e *Elastic) GetContractVersions(network, address string) ([]string, error) {
	b := boolQ(
		must(
			matchPhrase("network", network),
			matchPhrase("address", address),
		),
		should(
			term("vesting", true),
			matchPhrase("hash", ""),
		),
	)
	query := newQuery().Query(b).Sort("level", "desc").All()

	data, err := e.query([]string{DocMigrations}, query)
	if err != nil {
		return nil, err
	}

	versions := make([]string, 0)
	for _, hit := range data.Get("hits.hits").Array() {
		if hit.Get("_source.vesting").Bool() {
			versions = append(versions, "vesting")
		} else {
			versions = append(versions, hit.Get("_source.protocol").String())
		}
	}
	return versions, nil
}
