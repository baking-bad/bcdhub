package tokenbalance

import (
	"math/big"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
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
		),
	).Sort("token_id", "desc").All()

	size = core.GetSize(size, storage.es.MaxPageSize)

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
		),
	)
	count, err := storage.es.CountItems([]string{models.DocTokenBalances}, countQuery)
	if err != nil {
		return nil, 0, err
	}

	return tokenBalances, count, err
}

// Batch -
func (storage *Storage) Batch(network string, addresses []string) (map[string][]tokenbalance.TokenBalance, error) {
	if len(addresses) == 0 && len(addresses) > consts.DefaultSize {
		return nil, errors.Errorf("Invalid addresses count. Must be 0 < count < 10")
	}

	should := make([]core.Item, len(addresses))
	for i := range addresses {
		should[i] = core.MatchPhrase("address", addresses[i])
	}

	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(core.Term("network", network)),
			core.Should(should...),
			core.MinimumShouldMatch(1),
		),
	)
	var tokens []tokenbalance.TokenBalance
	if err := storage.es.GetAllByQuery(query, &tokens); err != nil {
		return nil, err
	}

	result := make(map[string][]tokenbalance.TokenBalance)
	for _, t := range tokens {
		if _, ok := result[t.Address]; !ok {
			result[t.Address] = []tokenbalance.TokenBalance{}
		}
		result[t.Address] = append(result[t.Address], t)
	}

	return result, nil
}

// NFTHolders -
func (storage *Storage) NFTHolders(network, contract string, tokenID int64) (tokens []tokenbalance.TokenBalance, err error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Term("network", network),
				core.MatchPhrase("contract", contract),
				core.Term("token_id", tokenID),
			),
			core.MustNot(
				core.Term("balance", "0"),
			),
		),
	)

	err = storage.es.GetAllByQuery(query, &tokens)
	return
}

type countByContractAgg struct {
	Aggs struct {
		Count struct {
			Buckets []core.Bucket `json:"buckets"`
		} `json:"count"`
	} `json:"aggregations"`
}

// CountByContract -
func (storage *Storage) CountByContract(network, address string) (map[string]int64, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Term("network", network),
				core.MatchPhrase("address", address),
			),
		),
	).Add(
		core.Aggs(
			core.AggItem{
				Name: "count",
				Body: core.TermsAgg("contract.keyword", core.MaxQuerySize),
			},
		),
	).Zero()

	var resp countByContractAgg
	if err := storage.es.Query([]string{models.DocTokenBalances}, query, &resp); err != nil {
		return nil, err
	}
	result := make(map[string]int64)
	for _, b := range resp.Aggs.Count.Buckets {
		result[b.Key] = b.DocCount
	}
	return result, nil
}

// GetTokenSupply -
func (storage *Storage) GetTokenSupply(network, contract string, tokenID int64) (string, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.Term("network", network),
				core.MatchPhrase("contract", contract),
				core.Term("token_id", tokenID),
			),
		),
	)

	var balances []tokenbalance.TokenBalance
	if err := storage.es.GetAllByQuery(query, &balances); err != nil {
		return "0", err
	}
	supply := new(big.Int)
	for i := range balances {
		supply.Add(supply, balances[i].Value)
	}
	return supply.String(), nil
}
