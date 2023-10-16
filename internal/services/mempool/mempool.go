package mempool

import (
	"context"

	"github.com/machinebox/graphql"
)

// urls
const (
	MempoolBaseURL = "https://mempool.dipdup.net/v1/graphql"
)

// Mempool -
type Mempool struct {
	client *graphql.Client
}

// NewMempool -
func NewMempool(url string) *Mempool {
	if url == "" {
		url = MempoolBaseURL
	}
	return &Mempool{
		client: graphql.NewClient(url),
	}
}

// Get -
func (m *Mempool) Get(ctx context.Context, address string) (result PendingOperations, err error) {
	req := graphql.NewRequest(`
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
	`)

	req.Var("address", address)
	req.Header.Set("Cache-Control", "no-cache")

	err = m.client.Run(ctx, req, &result)
	return
}

// GetByHash -
func (m *Mempool) GetByHash(ctx context.Context, hash string) (result PendingOperations, err error) {
	req := graphql.NewRequest(`
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
	`)

	req.Var("hash", hash)
	req.Header.Set("Cache-Control", "no-cache")

	err = m.client.Run(ctx, req, &result)
	return
}
