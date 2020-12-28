package operation

// Repository -
type Repository interface {
	GetByContract(network string, address string, size uint64, filters map[string]interface{}) (Pageable, error)
	GetStats(network, address string) (Stats, error)
	// Last - returns last operation. TODO: change network and address.
	Last(network string, address string, indexedTime int64) (Operation, error)

	// GetOperations - get operation by `filter`. `Size` - if 0 - return all, else certain `size` operations.
	// `Sort` - sort by time and content index by desc
	Get(filter map[string]interface{}, size int64, sort bool) ([]Operation, error)

	GetContract24HoursVolume(network, address string, entrypoints []string) (float64, error)
	GetTokensStats(network string, addresses, entrypoints []string) (map[string]TokenUsageStats, error)

	GetParticipatingContracts(network string, fromLevel int64, toLevel int64) ([]string, error)
	RecalcStats(network, address string) (ContractStats, error)
	GetDAppStats(string, []string, string) (DAppStats, error)
}
