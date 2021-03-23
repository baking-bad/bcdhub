package tokenbalance

import (
	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

// Update -
func (storage *Storage) Update(updates []*tokenbalance.TokenBalance) error {
	if len(updates) == 0 {
		return nil
	}
	buf := make([]tokenbalance.TokenBalance, 0)
	ids := make([]string, len(updates))
	for i := range updates {
		ids[i] = updates[i].GetID()
	}
	if err := storage.es.GetByIDs(&buf, ids...); err != nil {
		if !storage.es.IsRecordNotFound(err) {
			return err
		}
	}

	updatedModels := make([]models.Model, 0)
	insertedModels := make([]models.Model, 0)

	for i := range updates {
		if len(buf) == 0 {
			insertedModels = append(insertedModels, updates[i])
			continue
		}

		var found bool
		for j := range buf {
			if buf[j].GetID() == updates[i].GetID() {
				found = true
				updates[i].Sum(&buf[j])
				updatedModels = append(updatedModels, updates[i])
				break
			}
		}

		if !found {
			insertedModels = append(insertedModels, updates[i])
		}
	}

	if err := storage.es.BulkInsert(insertedModels); err != nil {
		return err
	}

	return storage.es.BulkUpdate(updatedModels)
}

// GetHolders -
func (storage *Storage) GetHolders(network, contract string, tokenID int64) ([]tokenbalance.TokenBalance, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Match("network", network),
				core.MatchPhrase("contract", contract),
				core.Term("token_id", tokenID),
			),
			core.MustNot(
				core.Term("balance", "0"),
			),
		),
	).All()

	balances := make([]tokenbalance.TokenBalance, 0)
	err := storage.es.GetAllByQuery(query, &balances)
	return balances, err
}

// GetAccountBalances -
func (storage *Storage) GetAccountBalances(network, address, contract string, size, offset int64) ([]tokenbalance.TokenBalance, int64, error) {
	filters := []core.Item{
		core.MatchPhrase("address", address),
		core.Match("network", network),
	}

	if contract != "" {
		filters = append(filters, core.MatchPhrase("contract", contract))
	}
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				filters...,
			),
			core.MustNot(
				core.Term("balance", "0"),
			),
		),
	).Sort("token_id", "desc").All()

	if size == 0 {
		size = consts.DefaultSize
	}

	tokenBalances := make([]tokenbalance.TokenBalance, 0)
	ctx := core.NewScrollContext(storage.es, query, size, consts.DefaultScrollSize)
	ctx.Offset = offset
	if err := ctx.Get(&tokenBalances); err != nil {
		return nil, 0, err
	}

	countQuery := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				filters...,
			),
			core.MustNot(
				core.Term("balance", "0"),
			),
		),
	)
	count, err := storage.es.CountItems([]string{models.DocTokenBalances}, countQuery)
	if err != nil {
		return nil, 0, err
	}

	return tokenBalances, count, err
}

// BurnNft -
func (storage *Storage) BurnNft(network, contract string, tokenID int64) error {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Term("network", network),
				core.MatchPhrase("contract", contract),
				core.Term("token_id", tokenID),
			),
		),
	)

	// 10 attempts in case of conflicts
	for i := 0; i < 10; i++ {
		response, err := storage.es.DeleteWithQuery([]string{models.DocTokenBalances}, query)
		if err != nil {
			return err
		}
		if response.VersionConflicts == 0 {
			break
		}
	}
	return nil
}
