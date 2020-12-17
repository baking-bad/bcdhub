package contract

// Repository -
type Repository interface {
	Get(map[string]interface{}) (Contract, error)
	GetMany(map[string]interface{}) ([]Contract, error)
	GetRandom() (Contract, error)
	GetMigrationsCount(string, string) (int64, error)
	GetAddressesByNetworkAndLevel(string, int64) ([]string, error)
	GetIDsByAddresses([]string, string) ([]string, error)
	GetByLevels(string, int64, int64) ([]string, error)
	IsFA(string, string) (bool, error)
	RecalcStats(string, string) (Stats, error)
	UpdateMigrationsCount(string, string) error
	GetDAppStats(string, []string, string) (DAppStats, error)
	GetByAddresses(addresses []Address) ([]Contract, error)
	GetTokens(string, string, int64, int64) ([]Contract, int64, error)
	GetProjectsLastContract() ([]Contract, error)
	GetSameContracts(Contract, int64, int64) (SameResponse, error)
	GetSimilarContracts(Contract, int64, int64) ([]Similar, int, error)
	GetDiffTasks() ([]DiffTask, error)
}
