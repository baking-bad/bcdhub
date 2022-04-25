package operation

import (
	"github.com/baking-bad/bcdhub/internal/models/account"
)

// Repository -
type Repository interface {
	GetByAccount(acc account.Account, size uint64, filters map[string]interface{}) (Pageable, error)
	// Last -  get last operation by `filters` with not empty deffated_storage.
	Last(filter map[string]interface{}, lastID int64) (Operation, error)
	GetByHash(hash string) ([]Operation, error)

	// GetOperations - get operation by `filter`. `Size` - if 0 - return all, else certain `size` operations.
	// `Sort` - sort by time and content index by desc
	Get(filter map[string]interface{}, size int64, sort bool) ([]Operation, error)

	GetContract24HoursVolume(address string, entrypoints []string) (float64, error)
	GetTokensStats(addresses, entrypoints []string) (map[string]TokenUsageStats, error)

	GetDAppStats(addresses []string, period string) (DAppStats, error)
	GetByIDs(ids ...int64) ([]Operation, error)
	GetByID(id int64) (Operation, error)
}
