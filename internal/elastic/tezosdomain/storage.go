package tezosdomain

import (
	"encoding/json"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tezosdomain"
	"github.com/pkg/errors"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

// ListDomains -
func (storage *Storage) ListDomains(network string, size, offset int64) (tezosdomain.DomainsResponse, error) {
	if size > consts.DefaultScrollSize {
		size = consts.DefaultScrollSize
	}

	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
			),
		),
	).Size(size).From(offset).Sort("timestamp", "desc")

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocTezosDomains}, query, &response); err != nil {
		return tezosdomain.DomainsResponse{}, err
	}
	if response.Hits.Total.Value == 0 {
		return tezosdomain.DomainsResponse{}, nil
	}

	domains := make([]tezosdomain.TezosDomain, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &domains[i]); err != nil {
			return tezosdomain.DomainsResponse{}, err
		}
	}
	return tezosdomain.DomainsResponse{
		Domains: domains,
		Total:   response.Hits.Total.Value,
	}, nil
}

// ResolveDomainByAddress -
func (storage *Storage) ResolveDomainByAddress(network string, address string) (*tezosdomain.TezosDomain, error) {
	if !helpers.IsAddress(address) {
		return nil, errors.Errorf("Invalid address: %s", address)
	}
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.MatchPhrase("address", address),
			),
		),
	).One()

	var response core.SearchResponse
	if err := storage.es.Query([]string{models.DocTezosDomains}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, core.NewRecordNotFoundError(models.DocTezosDomains, "")
	}

	var td tezosdomain.TezosDomain
	err := json.Unmarshal(response.Hits.Hits[0].Source, &td)
	return &td, err
}
