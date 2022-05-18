package elastic

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/search"
)

type sameContractsResponse struct {
	Hits struct {
		Total struct {
			Value    int64  `json:"value"`
			Relation string `json:"relation"`
		} `json:"total"`
		Hits []struct {
			Source search.Contract `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func newSameContractsQuery(contract contract.Contract, network string, networks ...string) Base {
	hash := contract.Babylon.Hash
	if contract.AlphaID > 0 && hash == "" {
		hash = contract.Alpha.Hash
	}
	return NewQuery().Query(
		Bool(
			Filter(
				Match("hash", hash),
				In("network", networks),
			),
			MustNot(
				Match("network", network),
				Match("address", contract.Account.Address),
			),
		),
	)
}

// SameContracts -
func (e *Elastic) SameContracts(contract contract.Contract, network string, networks []string, offset, size int64) (search.SameContracts, error) {
	if size == 0 {
		size = defaultSize
	}
	query := newSameContractsQuery(contract, network, networks...).Size(size).From(offset)

	var response sameContractsResponse
	if err := e.query([]string{models.DocContracts}, query, &response); err != nil {
		return search.SameContracts{}, err
	}

	same := search.SameContracts{
		Contracts: make([]search.Contract, 0),
	}

	for i := range response.Hits.Hits {
		same.Contracts = append(same.Contracts, response.Hits.Hits[i].Source)
	}

	if response.Hits.Total.Relation == "eq" {
		same.Count = response.Hits.Total.Value
	} else {
		count, err := e.count([]string{models.DocContracts}, newSameContractsQuery(contract, network))
		if err != nil {
			return search.SameContracts{}, err
		}
		same.Count = count
	}

	return same, nil
}
