package elastic

import (
	"io"

	"github.com/baking-bad/bcdhub/internal/elastic/search"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/mq"
)

// Model -
type Model interface {
	mq.IMessage

	GetID() string
	GetIndex() string
}

// IGeneral -
type IGeneral interface {
	CreateIndexes() error
	DeleteIndices(indices []string) error
	DeleteByLevelAndNetwork([]string, string, int64) error
	DeleteByContract(indices []string, network, address string) error
	GetAll(interface{}) error
	GetByID(Model) error
	GetByIDs(output interface{}, ids ...string) error
	GetByNetwork(string, interface{}) error
	GetByNetworkWithSort(string, string, string, interface{}) error
	UpdateDoc(model Model) (err error)
	UpdateFields(string, string, interface{}, ...string) error
}

// IBalanceUpdate -
type IBalanceUpdate interface {
	GetBalance(network, address string) (int64, error)
}

// IBigMap -
type IBigMap interface {
	GetBigMapKey(network, keyHash string, ptr int64) (BigMapDiff, error)
	GetBigMapKeys(ctx GetBigMapKeysContext) ([]BigMapDiff, error)
	GetBigMapsForAddress(string, string) ([]models.BigMapDiff, error)
	GetBigMapHistory(int64, string) ([]models.BigMapAction, error)
	GetBigMapValuesByKey(string) ([]BigMapDiff, error)
}

// IBigMapDiff -
type IBigMapDiff interface {
	GetBigMapDiffsForAddress(string) ([]models.BigMapDiff, error)
	GetBigMapDiffsPrevious([]models.BigMapDiff, int64, string) ([]models.BigMapDiff, error)
	GetBigMapDiffsUniqueByOperationID(string) ([]models.BigMapDiff, error)
	GetBigMapDiffsByPtrAndKeyHash(int64, string, string, int64, int64) ([]BigMapDiff, int64, error)
	GetBigMapDiffsByOperationID(string) ([]*models.BigMapDiff, error)
	GetBigMapDiffsByPtr(string, string, int64) ([]models.BigMapDiff, error)
}

// IBlock -
type IBlock interface {
	GetBlock(string, int64) (models.Block, error)
	GetLastBlock(string) (models.Block, error)
	GetLastBlocks() ([]models.Block, error)
	GetNetworkAlias(chainID string) (string, error)
}

// IBulk -
type IBulk interface {
	BulkInsert([]Model) error
	BulkUpdate([]Model) error
	BulkDelete([]Model) error
	BulkRemoveField(string, []Model) error
	BulkUpdateField(where []models.Contract, fields ...string) error
}

// IContract -
type IContract interface {
	GetContract(map[string]interface{}) (models.Contract, error)
	GetContractRandom() (models.Contract, error)
	GetContractMigrationStats(string, string) (ContractMigrationsStats, error)
	GetContractAddressesByNetworkAndLevel(string, int64) ([]string, error)
	GetContracts(map[string]interface{}) ([]models.Contract, error)
	GetContractsIDByAddress([]string, string) ([]string, error)
	GetAffectedContracts(string, int64, int64) ([]string, error)
	IsFAContract(string, string) (bool, error)
	RecalcContractStats(string, string) (ContractStats, error)
	UpdateContractMigrationsCount(string, string) error
	GetDAppStats(string, []string, string) (DAppStats, error)
	GetContractsByAddresses(addresses []Address) ([]models.Contract, error)
}

// IDomains -
type IDomains interface {
	ListDomains(network string, size, offset int64) (DomainsResponse, error)
	ResolveDomainByAddress(network string, address string) (*models.TezosDomain, error)
}

// IEvents -
type IEvents interface {
	GetEvents([]SubscriptionRequest, int64, int64) ([]Event, error)
}

// IMigrations -
type IMigrations interface {
	GetMigrations(string, string) ([]models.Migration, error)
}

// IOperations -
type IOperations interface {
	GetOperationsForContract(string, string, uint64, map[string]interface{}) (PageableOperations, error)
	GetLastOperation(string, string, int64) (models.Operation, error)
	GetOperationsStats(network, address string) (OperationsStats, error)

	// GetOperations - get operation by `filter`. `Size` - if 0 - return all, else certain `size` operations.
	// `Sort` - sort by time and content index by desc
	GetOperations(filter map[string]interface{}, size int64, sort bool) ([]models.Operation, error)
}

// IProjects -
type IProjects interface {
	GetProjectsLastContract() ([]models.Contract, error)
	GetSameContracts(models.Contract, int64, int64) (SameContractsResponse, error)
	GetSimilarContracts(models.Contract, int64, int64) ([]SimilarContract, int, error)
	GetDiffTasks() ([]DiffTask, error)
}

// IProtocol -
type IProtocol interface {
	GetProtocol(string, string, int64) (models.Protocol, error)
	GetSymLinks(string, int64) (map[string]struct{}, error)
}

// ISearch -
type ISearch interface {
	SearchByText(string, int64, []string, map[string]interface{}, bool) (search.Result, error)
}

// ISnapshot -
type ISnapshot interface {
	CreateAWSRepository(string, string, string) error
	ListRepositories() ([]Repository, error)
	CreateSnapshots(string, string, []string) error
	RestoreSnapshots(string, string, []string) error
	ListSnapshots(string) (string, error)
	SetSnapshotPolicy(string, string, string, string, int64) error
	GetAllPolicies() ([]string, error)
	GetMappings([]string) (map[string]string, error)
	CreateMapping(string, io.Reader) error
	ReloadSecureSettings() error
}

// IStats -
type IStats interface {
	GetNetworkCountStats(string) (map[string]int64, error)
	GetDateHistogram(period string, opts ...HistogramOption) ([][]int64, error)
	GetCallsCountByNetwork() (map[string]int64, error)
	GetContractStatsByNetwork() (map[string]ContractCountStats, error)
	GetFACountByNetwork() (map[string]int64, error)
	GetLanguagesForNetwork(network string) (map[string]int64, error)
}

// ITokens -
type ITokens interface {
	GetTokens(string, string, int64, int64) ([]models.Contract, int64, error)
	GetTokensStats(string, []string, []string) (map[string]TokenUsageStats, error)
	GetTokenVolumeSeries(string, string, []string, []tzip.DAppContract, uint) ([][]int64, error)
	GetBalances(string, string, int64, ...TokenBalance) (map[TokenBalance]int64, error)
	GetAccountBalances(string, string) ([]models.TokenBalance, error)
	GetTokenSupply(network, address string, tokenID int64) (result TokenSupply, err error)
	GetTransfers(ctx GetTransfersContext) (TransfersResponse, error)
	GetAllTransfers(network string, level int64) ([]models.Transfer, error)
	UpdateTokenBalances(updates []*models.TokenBalance) error
	GetHolders(network, contract string, tokenID int64) ([]models.TokenBalance, error)
}

// ITZIP -
type ITZIP interface {
	GetTZIP(network, address string) (models.TZIP, error)
	GetTZIPWithEvents() ([]models.TZIP, error)
	GetTokenMetadata(ctx GetTokenMetadataContext) ([]TokenMetadata, error)
	GetDApps() ([]tzip.DApp, error)
	GetDAppBySlug(slug string) (*tzip.DApp, error)
	GetBySlug(slug string) (*models.TZIP, error)
	GetAliases(network string) ([]models.TZIP, error)
	GetAliasesMap(network string) (map[string]string, error)
	GetAlias(network, address string) (*models.TZIP, error)
}

// IElastic -
type IElastic interface {
	IGeneral
	IBalanceUpdate
	IBigMap
	IBigMapDiff
	IBlock
	IBulk
	IContract
	IDomains
	IEvents
	IMigrations
	IOperations
	IProjects
	IProtocol
	ISearch
	ISnapshot
	IStats
	ITokens
	ITZIP
}
