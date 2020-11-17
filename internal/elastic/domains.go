package elastic

import (
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

// ListDomains -
func (e *Elastic) ListDomains(network string, size, offset int64) ([]models.TezosDomain, error) {
	if size > defaultScrollSize {
		size = defaultScrollSize
	}

	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
			),
		),
	).Size(size).From(offset).Sort("timestamp", "desc")

	var response SearchResponse
	if err := e.query([]string{DocTezosDomains}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, nil
	}

	domains := make([]models.TezosDomain, len(response.Hits.Hits))
	for i := range response.Hits.Hits {
		if err := json.Unmarshal(response.Hits.Hits[i].Source, &domains[i]); err != nil {
			return nil, err
		}
	}
	return domains, nil
}

// ResolveDomainByAddress -
func (e *Elastic) ResolveDomainByAddress(network string, address string) (*models.TezosDomain, error) {
	if !helpers.IsAddress(address) {
		return nil, errors.Errorf("Invalid address: %s", address)
	}
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				matchPhrase("address", address),
			),
		),
	).One()

	var response SearchResponse
	if err := e.query([]string{DocTezosDomains}, query, &response); err != nil {
		return nil, err
	}
	if response.Hits.Total.Value == 0 {
		return nil, NewRecordNotFoundError(DocTezosDomains, "", query)
	}

	var td models.TezosDomain
	err := json.Unmarshal(response.Hits.Hits[0].Source, &td)
	return &td, err
}
