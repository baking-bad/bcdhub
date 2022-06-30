package domains

import (
	"github.com/baking-bad/bcdhub/internal/models/contract"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

// Repository -
type Repository interface {
	TokenBalances(contract string, accountID int64, size, offset int64, sort string, hideZeroBalances bool) (TokenBalanceResponse, error)
	Transfers(ctx transfer.GetContext) (TransfersResponse, error)
	BigMapDiffs(lastID, size int64) ([]BigMapDiff, error)

	Same(network string, c contract.Contract, limit, offset int) ([]Same, error)
	SameCount(c contract.Contract) (int, error)
}
