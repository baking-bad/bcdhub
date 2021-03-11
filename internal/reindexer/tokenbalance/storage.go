package tokenbalance

import (
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/reindexer/core"
	"github.com/restream/reindexer"
)

// Storage -
type Storage struct {
	db *core.Reindexer
}

// NewStorage -
func NewStorage(db *core.Reindexer) *Storage {
	return &Storage{db}
}

// const scriptUpdateBalance = `{"source": "ctx._source.balance = ctx._source.balance + (long)params.delta", "lang": "painless", "params": { "delta": %d }}`

// Update -
func (storage *Storage) Update(updates []*tokenbalance.TokenBalance) error {
	if len(updates) == 0 {
		return nil
	}

	tx, err := storage.db.BeginTx(models.DocTokenBalances)
	for err != nil {
		return err
	}

	// bulk := bytes.NewBuffer([]byte{})
	// for i := range updates {
	// 	bulk.WriteString(fmt.Sprintf(`{ "update": { "_id": "%s"}}`, updates[i].GetID()))
	// 	bulk.WriteByte('\n')

	// 	script := fmt.Sprintf(scriptUpdateBalance, updates[i].Balance)

	// 	upsert, err := json.Marshal(updates[i])
	// 	if err != nil {
	// 		return err
	// 	}

	// 	bulk.WriteString(fmt.Sprintf(`{ "script": %s, "upsert": %s }`, script, string(upsert)))
	// 	bulk.WriteByte('\n')
	// 	if (i%1000 == 0 && i > 0) || i == len(updates)-1 {
	// 		if err := storage.bulkUpsertBalances(bulk); err != nil {
	// 			return err
	// 		}
	// 		bulk.Reset()
	// 	}
	// }
	return tx.Commit()
}

// func (storage *Storage) bulkUpsertBalances(buf *bytes.Buffer) error {
// 	req := esapi.BulkRequest{
// 		Body:    bytes.NewReader(buf.Bytes()),
// 		Refresh: "true",
// 		Index:   models.DocTokenBalances,
// 	}

// 	res, err := req.Do(context.Background(), storage.es)
// 	if err != nil {
// 		return err
// 	}
// 	defer res.Body.Close()

// 	var response core.BulkResponse
// 	return storage.es.GetResponse(res, &response)
// }

// GetHolders -
func (storage *Storage) GetHolders(network, contract string, tokenID int64) (balances []tokenbalance.TokenBalance, err error) {
	query := storage.db.Query(models.DocTokenBalances).
		Match("network", network).
		Match("contract", contract).
		WhereInt64("token_id", reindexer.EQ, tokenID).
		WhereInt64("balance", reindexer.GT, 0)

	err = storage.db.GetAllByQuery(query, &balances)
	return
}

// GetAccountBalances -
func (storage *Storage) GetAccountBalances(network, address string) (tokenBalances []tokenbalance.TokenBalance, err error) {
	query := storage.db.Query(models.DocTokenBalances).
		Match("network", network).
		Match("address", address)

	err = storage.db.GetAllByQuery(query, &tokenBalances)
	return
}

// BurnNft -
func (storage *Storage) BurnNft(network, contract string, tokenID int64) error {
	return nil
}
