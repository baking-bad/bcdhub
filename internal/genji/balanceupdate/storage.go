package balanceupdate

import (
	"github.com/baking-bad/bcdhub/internal/genji/core"
	"github.com/baking-bad/bcdhub/internal/models"
)

// Storage -
type Storage struct {
	db *core.Genji
}

// NewStorage -
func NewStorage(db *core.Genji) *Storage {
	return &Storage{db}
}

// GetBalance -
// TODO: get balance
func (storage *Storage) GetBalance(network, address string) (int64, error) {
	builder := core.NewBuilder()

	builder.Select(models.DocBalanceUpdates, "*").And(
		core.NewEq("network", network),
		core.NewEq("contract", address),
	)
	// query := core.NewQuery().Query(
	// 	core.Bool(
	// 		core.Filter(
	// 			core.Match("network", network),
	// 			core.MatchPhrase("contract", address),
	// 		),
	// 	),
	// ).Add(
	// 	core.Aggs(
	// 		core.AggItem{
	// 			Name: "balance",
	// 			Body: core.Sum("change"),
	// 		},
	// 	),
	// ).Zero()

	res, err := storage.db.Query(builder.String())
	if err != nil {
		return 0, err
	}
	defer res.Close()

	return 0, nil
}
