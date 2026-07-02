package mempool

import (
	"bytes"
	"context"
	"net/http"
	"time"

	stdJSON "encoding/json"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// urls
const (
	MempoolBaseURL = "https://mempool.dipdup.net/v1/graphql"
)

// Mempool -
type Mempool struct {
	url    string
	client *http.Client
}

// NewMempool -
func NewMempool(url string) *Mempool {
	if url == "" {
		url = MempoolBaseURL
	}
	return &Mempool{
		url: url,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type graphqlRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type graphqlResponse struct {
	Data   stdJSON.RawMessage `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (m *Mempool) run(ctx context.Context, query string, variables map[string]interface{}, result interface{}) error {
	body, err := json.Marshal(graphqlRequest{
		Query:     query,
		Variables: variables,
	})
	if err != nil {
		return errors.Wrap(err, "encoding mempool request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Accept", "application/json; charset=utf-8")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// mempool service is not deployed for the network: treat as empty mempool
	if resp.StatusCode == http.StatusNotFound {
		return nil
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("mempool service responded with %s", resp.Status)
	}

	var gr graphqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return errors.Wrap(err, "decoding mempool response")
	}
	if len(gr.Errors) > 0 {
		return errors.Errorf("mempool service: %s", gr.Errors[0].Message)
	}
	if len(gr.Data) == 0 {
		return errors.New("mempool service: empty data in response")
	}
	return json.Unmarshal(gr.Data, result)
}

// Get -
func (m *Mempool) Get(ctx context.Context, address string) (result PendingOperations, err error) {
	query := `
		query ($address: String!) {
			originations(where: {source: {_eq: $address}, _not: {status: {_eq: "in_chain"}}}) {
				balance
				branch
				created_at
				delegate
				errors
				expiration_level
				fee
				gas_limit
				kind
				level
				signature
				source
				status
				storage
				storage_limit
				updated_at
				network
				hash
				raw
				protocol
			}
			transactions(
				where: {_or: [{source: {_eq: $address}}, {destination: {_eq: $address}}], _not: {status: {_eq: "in_chain"}}}
			) {
				amount
				branch
				created_at
				errors
				expiration_level
				fee
				gas_limit
				kind
				level
				parameters
				signature
				source
				status
				storage_limit
				updated_at
				destination
				network
				hash
				counter
				raw
				protocol
			}
		}
	`

	err = m.run(ctx, query, map[string]interface{}{"address": address}, &result)
	return
}

// GetByHash -
func (m *Mempool) GetByHash(ctx context.Context, hash string) (result PendingOperations, err error) {
	query := `
		query ($hash: String!) {
			originations(where: {hash: {_eq: $hash}, _not: {status: {_eq: "in_chain"}}}) {
				balance
				branch
				created_at
				delegate
				errors
				expiration_level
				fee
				gas_limit
				kind
				level
				signature
				source
				status
				storage
				storage_limit
				updated_at
				network
				hash
				raw
				protocol
			}
			transactions(where: {hash: {_eq: $hash}, _not: {status: {_eq: "in_chain"}}}) {
				amount
				branch
				created_at
				errors
				expiration_level
				fee
				gas_limit
				kind
				level
				parameters
				signature
				source
				status
				storage_limit
				updated_at
				destination
				network
				hash
				counter
				raw
				protocol
			}
		}
	`

	err = m.run(ctx, query, map[string]interface{}{"hash": hash}, &result)
	return
}
