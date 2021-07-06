package domains

import (
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// Repository -
type Repository interface {
	TokenBalances(network types.Network, contract, address string, size, offset int64, sort string) (TokenBalanceResponse, error)
	Transfers(ctx transfer.GetContext) (TransfersResponse, error)
	BigMapDiffs(lastID, size int64) ([]BigMapDiff, error)
}
