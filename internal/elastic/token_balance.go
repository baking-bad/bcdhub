package elastic

import (
	"bytes"
	"context"
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

const scriptUpdateBalance = `{"source": "ctx._source.balance = ctx._source.balance + params.delta", "lang": "painless", "params": { "delta": %d }}`

// UpdateTokenBalances -
func (e *Elastic) UpdateTokenBalances(updates []*models.TokenBalance) error {
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

		bulk.WriteString(fmt.Sprintf(`{ "script": %s, "scripted_upsert": true,  "upsert": %s }`, script, string(upsert)))
		bulk.WriteByte('\n')
		if (i%1000 == 0 && i > 0) || i == len(updates)-1 {
			if err := e.bulkUpsertBalances(bulk); err != nil {
				return err
			}
			bulk.Reset()
		}
	}
	return nil
}

func (e *Elastic) bulkUpsertBalances(buf *bytes.Buffer) error {
	req := esapi.BulkRequest{
		Body:    bytes.NewReader(buf.Bytes()),
		Refresh: "true",
		Index:   DocTokenBalances,
	}

	res, err := req.Do(context.Background(), e)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	var response BulkResponse
	return e.getResponse(res, &response)
}

// GetHolders -
func (e *Elastic) GetHolders(network, contract string, tokenID int64) ([]models.TokenBalance, error) {
	query := newQuery().Query(
		boolQ(
			filter(
				matchQ("network", network),
				matchPhrase("contract", contract),
				term("token_id", tokenID),
			),
		),
	).All()

	balances := make([]models.TokenBalance, 0)
	err := e.getAllByQuery(query, &balances)
	return balances, err
}
