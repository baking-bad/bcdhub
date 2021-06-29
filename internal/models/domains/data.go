package domains

import (
	"github.com/baking-bad/bcdhub/internal/models/bigmap"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/protocol"
	"github.com/baking-bad/bcdhub/internal/models/tokenmetadata"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

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

// Transfer -
type Transfer struct {
	*transfer.Transfer
	Hash     string
	Symbol   string
	Name     string
	Counter  int64
	Nonce    *int64
	Decimals *int64
}

// TransfersResponse -
type TransfersResponse struct {
	Total     int64
	LastID    string
	Transfers []Transfer
}

// BigMapDiff -
type BigMapDiff struct {
	*bigmap.Diff

	Operation *operation.Operation
	Protocol  *protocol.Protocol
}

// TableName -
func (BigMapDiff) TableName() string {
	return "big_map_diffs"
}
