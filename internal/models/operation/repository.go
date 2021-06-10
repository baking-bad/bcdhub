package operation

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	GetByContract(network types.Network, address string, size uint64, filters map[string]interface{}) (Pageable, error)
	GetStats(network types.Network, address string) (Stats, error)
	// Last - returns last operation. TODO: change network and address.
	Last(network types.Network, address string, indexedTime int64) (Operation, error)

	// GetOperations - get operation by `filter`. `Size` - if 0 - return all, else certain `size` operations.
	// `Sort` - sort by time and content index by desc
	Get(filter map[string]interface{}, size int64, sort bool) ([]Operation, error)

	GetContract24HoursVolume(network types.Network, address string, entrypoints []string) (float64, error)
	GetTokensStats(network types.Network, addresses, entrypoints []string) (map[string]TokenUsageStats, error)

	GetParticipatingContracts(network types.Network, fromLevel int64, toLevel int64) ([]string, error)
	GetDAppStats(network types.Network, addresses []string, period string) (DAppStats, error)
	GetByIDs(ids ...int64) ([]Operation, error)
}
