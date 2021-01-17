package tokenbalance

import (
	"bytes"
	stdJSON "encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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
		return err
	}

	updatedModels := make([]models.Model, 0)
	for i := range updates {
		for j := range buf {
			if buf[j].GetID() == updates[i].GetID() {
				updates[i].Sum(&buf[j])
				updatedModels = append(updatedModels, updates[i])
				break
			}
		}
	}

	return storage.es.BulkUpdate(updatedModels)
}

func (storage *Storage) updateBalances(items []*tokenbalance.TokenBalance) error {
	bulk := bytes.NewBuffer([]byte{})
	for i := range items {
		meta := fmt.Sprintf(`{"index":{"_id":"%s","_index":"%s"}}`, items[i].GetID(), items[i].GetIndex())
		if _, err := bulk.WriteString(meta); err != nil {
			return err
		}

		if err := bulk.WriteByte('\n'); err != nil {
			return err
		}

		data, err := json.Marshal(items[i])
		if err != nil {
			return err
		}

		if err := stdJSON.Compact(bulk, data); err != nil {
			return err
		}
		if err := bulk.WriteByte('\n'); err != nil {
			return err
		}

		if (i%1000 == 0 && i > 0) || i == len(items)-1 {
			if err := storage.es.Bulk(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
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
func (storage *Storage) GetAccountBalances(network, address string) ([]tokenbalance.TokenBalance, error) {
	query := core.NewQuery().Query(
		core.Bool(
			core.Filter(
				core.MatchPhrase("address", address),
				core.Match("network", network),
			),
		),
	).All()

	tokenBalances := make([]tokenbalance.TokenBalance, 0)
	err := storage.es.GetAllByQuery(query, &tokenBalances)
	return tokenBalances, err
}
