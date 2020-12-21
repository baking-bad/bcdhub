package balanceupdate

import (
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

type getBalanceResponse struct {
	Agg struct {
		Balance core.FloatValue `json:"balance"`
	} `json:"aggregations"`
}

// GetBalance -
func (storage *Storage) GetBalance(network, address string) (int64, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.MatchPhrase("contract", address),
			),
		),
	).Add(
		core.Aggs(
			core.AggItem{
				Name: "balance",
				Body: core.Sum("change"),
			},
		),
	).Zero()

	var response getBalanceResponse
	if err := storage.es.Query([]string{models.DocBalanceUpdates}, query, &response); err != nil {
		return 0, err
	}
	return int64(response.Agg.Balance.Value), nil
}
