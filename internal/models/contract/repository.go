package contract

// Repository -
type Repository interface {
	Get(by map[string]interface{}) (Contract, error)
	GetMany(by map[string]interface{}) ([]Contract, error)
	GetRandom(network string) (Contract, error)
	GetAddressesByNetworkAndLevel(network string, maxLevel int64) ([]string, error)
	GetIDsByAddresses(addresses []string, network string) ([]string, error)
	IsFA(network, address string) (bool, error)
	UpdateMigrationsCount(address, network string) error
	GetByAddresses(addresses []Address) ([]Contract, error)
	GetTokens(network, tokenInterface string, offset, size int64) ([]Contract, int64, error)
	GetProjectsLastContract(contract *Contract) ([]Contract, error)
	GetSameContracts(contact Contract, manager string, size, offset int64) (SameResponse, error)
	GetSimilarContracts(Contract, int64, int64) ([]Similar, int, error)
	GetDiffTasks() ([]DiffTask, error)
	UpdateField(where []Contract, fields ...string) error
}
