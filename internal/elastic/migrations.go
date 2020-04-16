package elastic

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
)

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

// GetAllMigrations -
func (e *Elastic) GetAllMigrations(network string) ([]models.Migration, error) {
	migrations := make([]models.Migration, 0)

	query := newQuery().Query(
		boolQ(
			must(
				matchQ("network", network),
			),
		),
	).Sort("level", "asc")
	result, err := e.createScroll(DocMigrations, 1000, query)
	if err != nil {
		return nil, err
	}
	for {
		scrollID := result.Get("_scroll_id").String()
		hits := result.Get("hits.hits")
		if hits.Get("#").Int() < 1 {
			break
		}

		for _, hit := range hits.Array() {
			var migration models.Migration
			migration.ParseElasticJSON(hit)
			migrations = append(migrations, migration)
		}

		result, err = e.queryScroll(scrollID)
		if err != nil {
			return nil, err
		}
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
			matchQ("hash", ""),
		),
		minimumShouldMatch(1),
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

// GetMigrationByID -
func (e *Elastic) GetMigrationByID(id string) (migration models.Migration, err error) {
	resp, err := e.GetByID(DocMigrations, id)
	if err != nil {
		return
	}
	if !resp.Get("found").Bool() {
		return migration, fmt.Errorf("Unknown migration with ID %s", id)
	}
	migration.ParseElasticJSON(resp)
	return
}
