package domains

import (
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

// Repository -
type Repository interface {
	TokenBalances(contract string, accountID int64, size, offset int64, sort string, hideZeroBalances bool) (TokenBalanceResponse, error)
	Transfers(ctx transfer.GetContext) (TransfersResponse, error)
	BigMapDiffs(lastID, size int64) ([]BigMapDiff, error)
}
