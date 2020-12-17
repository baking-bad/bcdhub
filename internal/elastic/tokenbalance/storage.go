package tokenbalance

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/elastic/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Storage -
type Storage struct {
	es *core.Elastic
}

// NewStorage -
func NewStorage(es *core.Elastic) *Storage {
	return &Storage{es}
}

const scriptUpdateBalance = `{"source": "ctx._source.balance = ctx._source.balance + (long)params.delta", "lang": "painless", "params": { "delta": %d }}`

// UpdateTokenBalances -
func (storage *Storage) UpdateTokenBalances(updates []*tokenbalance.TokenBalance) error {
	if len(updates) == 0 {
		return nil
	}
	bulk := bytes.NewBuffer([]byte{})
	for i := range updates {
		bulk.WriteString(fmt.Sprintf(`{ "update": { "_id": "%s"}}`, updates[i].GetID()))
		bulk.WriteByte('\n')

		script := fmt.Sprintf(scriptUpdateBalance, updates[i].Balance)

		upsert, err := json.Marshal(updates[i])
		if err != nil {
			return err
		}

		bulk.WriteString(fmt.Sprintf(`{ "script": %s, "upsert": %s }`, script, string(upsert)))
		bulk.WriteByte('\n')
		if (i%1000 == 0 && i > 0) || i == len(updates)-1 {
			if err := storage.bulkUpsertBalances(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

func (storage *Storage) bulkUpsertBalances(buf *bytes.Buffer) error {
	req := esapi.BulkRequest{
		Body:    bytes.NewReader(buf.Bytes()),
		Refresh: "true",
		Index:   consts.DocTokenBalances,
	}

	res, err := req.Do(context.Background(), storage.es)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var response core.BulkResponse
	return storage.es.GetResponse(res, &response)
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
				core.Term("balance", 0),
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
