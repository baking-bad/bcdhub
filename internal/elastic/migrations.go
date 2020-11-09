package elastic

import (
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

	var response SearchResponse
	if err := e.query([]string{DocMigrations}, query, &response); err != nil {
		return nil, err
	}

	migrations := make([]models.Migration, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &migrations[i]); err != nil {
			return nil, err
		}
	}
	return migrations, nil
}
