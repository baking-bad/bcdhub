package contract

// Repository -
type Repository interface {
	Get(map[string]interface{}) (Contract, error)
	GetMany(map[string]interface{}) ([]Contract, error)
	GetRandom() (Contract, error)
	GetAddressesByNetworkAndLevel(string, int64) ([]string, error)
	GetIDsByAddresses([]string, string) ([]string, error)
	IsFA(string, string) (bool, error)
	UpdateMigrationsCount(string, string) error
	GetByAddresses(addresses []Address) ([]Contract, error)
	GetTokens(string, string, int64, int64) ([]Contract, int64, error)
	GetProjectsLastContract() ([]Contract, error)
	GetSameContracts(Contract, int64, int64) (SameResponse, error)
	GetSimilarContracts(Contract, int64, int64) ([]Similar, int, error)
	GetDiffTasks() ([]DiffTask, error)
}
