package tzip

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
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
func (storage *Storage) Get(network, address string) (t tzip.TZIP, err error) {
	t.Address = address
	t.Network = network
	err = storage.es.GetByID(&t)
	return
}

// GetDApps -
func (storage *Storage) GetDApps() ([]tzip.DApp, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Exists("dapps"),
			),
		),
	).Sort("dapps.order", "asc").All()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocTZIP}, query, &response, "dapps"); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, core.NewRecordNotFoundError(models.DocTZIP, "")
	}

	tokens := make([]tzip.DApp, 0)
	for _, hit := range response.Hits.Hits {
		var model tzip.TZIP
		if err := json.Unmarshal(hit.Source, &model); err != nil {
			return nil, err
		}
		tokens = append(tokens, model.DApps...)
	}

	return tokens, nil
}

// GetDAppBySlug -
func (storage *Storage) GetDAppBySlug(slug string) (*tzip.DApp, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("dapps.slug", slug),
			),
		),
	).One()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocTZIP}, query, &response, "dapps"); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, core.NewRecordNotFoundError(models.DocTZIP, "")
	}

	var model tzip.TZIP
	if err := json.Unmarshal(response.Hits.Hits[0].Source, &model); err != nil {
		return nil, err
	}
	return &model.DApps[0], nil
}

// GetBySlug -
func (storage *Storage) GetBySlug(slug string) (*tzip.TZIP, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Term("slug.keyword", slug),
			),
		),
	).One()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocTZIP}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, core.NewRecordNotFoundError(models.DocTZIP, "")
	}

	var data tzip.TZIP
	err := json.Unmarshal(response.Hits.Hits[0].Source, &data)
	return &data, err
}

// GetAliasesMap -
func (storage *Storage) GetAliasesMap(network string) (map[string]string, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
			),
		),
	).All()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocTZIP}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, core.NewRecordNotFoundError(models.DocTZIP, "")
	}

	aliases := make(map[string]string)
	for _, hit := range response.Hits.Hits {
		var data tzip.TZIP
		if err := json.Unmarshal(hit.Source, &data); err != nil {
			return nil, err
		}
		aliases[data.Address] = data.Name
	}

	return aliases, nil
}

// GetAliases -
func (storage *Storage) GetAliases(network string) ([]tzip.TZIP, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.Exists("name"),
			),
		),
	).All()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocTZIP}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, core.NewRecordNotFoundError(models.DocTZIP, "")
	}

	aliases := make([]tzip.TZIP, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &aliases[i]); err != nil {
			return nil, err
		}
	}
	return aliases, nil
}

// GetAlias -
func (storage *Storage) GetAlias(network, address string) (*tzip.TZIP, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.MatchPhrase("address", address),
			),
		),
	).One()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocTZIP}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, core.NewRecordNotFoundError(models.DocTZIP, "")
	}

	var data tzip.TZIP
	err := json.Unmarshal(response.Hits.Hits[0].Source, &data)
	return &data, err
}

// GetWithEvents -
func (storage *Storage) GetWithEvents() ([]tzip.TZIP, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Exists("events"),
			),
		),
	).All()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocTZIP}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, core.NewRecordNotFoundError(models.DocTZIP, "")
	}

	tokens := make([]tzip.TZIP, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &tokens[i]); err != nil {
			return nil, err
		}
	}
	return tokens, nil
}

// GetWithEventsCounts -
func (storage *Storage) GetWithEventsCounts() (int64, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Exists("events"),
			),
		),
	)

	return storage.es.CountItems([]string{models.DocTZIP}, query)
}
