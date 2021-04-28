package domains

import "github.com/baking-bad/bcdhub/internal/models/tokenmetadata"

// TokenBalance -
type TokenBalance struct {
	tokenmetadata.TokenMetadata

	Balance string
}

// TokenBalanceResponse -
type TokenBalanceResponse struct {
	Balances []TokenBalance
	Count    int64
}
