package elastic

import (
	"io"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/mq"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/tidwall/gjson"
)

// Model -
type Model interface {
	mq.IQueued

	GetID() string
	GetIndex() string
	ParseElasticJSON(gjson.Result)
}

// Scorable -
type Scorable interface {
	GetScores(string) []string
	FoundByName(gjson.Result) string
}

// IGeneral -
type IGeneral interface {
	AddDocument(interface{}, string) (string, error)
	AddDocumentWithID(interface{}, string, string) (string, error)
	CreateIndexIfNotExists(string) error
	CreateIndexes() error
	DeleteByLevelAndNetwork([]string, string, int64) error
	GetAll(interface{}) error
	GetAPI() *esapi.API
	GetByID(Model) error
	GetByNetwork(string, interface{}) error
	GetByNetworkWithSort(string, string, string, interface{}) error
	UpdateDoc(string, string, interface{}) (gjson.Result, error)
	UpdateFields(string, string, interface{}, ...string) error
}

// IBigMap -
type IBigMap interface {
	GetBigMapKeys(int64, string, string, int64, int64) ([]BigMapDiff, error)
	GetBigMapsForAddress(string, string) ([]models.BigMapDiff, error)
	GetBigMapHistory(int64, string) ([]models.BigMapAction, error)
}

// IBigMapDiff -
type IBigMapDiff interface {
	GetBigMapDiffsForAddress(string) ([]models.BigMapDiff, error)
	GetBigMapDiffsPrevious([]models.BigMapDiff, int64, string) ([]models.BigMapDiff, error)
	GetBigMapDiffsUniqueByOperationID(string) ([]models.BigMapDiff, error)
	GetBigMapDiffsByPtrAndKeyHash(int64, string, string, int64, int64) ([]BigMapDiff, int64, error)
	GetBigMapDiffsJSONByOperationID(string) ([]gjson.Result, error)
	GetBigMapDiffsByPtr(string, string, int64) ([]models.BigMapDiff, error)
	GetBigMapDiffsWithEmptyPtr() ([]models.BigMapDiff, error)
}

// IBlock -
type IBlock interface {
	GetBlock(string, int64) (models.Block, error)
	GetLastBlock(string) (models.Block, error)
	GetLastBlocks() ([]models.Block, error)
}

// IBulk -
type IBulk interface {
	BulkInsert([]Model) error
	BulkUpdate([]Model) error
	BulkDelete([]Model) error
	BulkRemoveField(string, []Model) error
}

// IContract -
type IContract interface {
	GetContract(map[string]interface{}) (models.Contract, error)
	GetContractRandom() (models.Contract, error)
	GetContractWithdrawn(string, string) (int64, error)
	GetContractMigrationStats(string, string) (ContractMigrationsStats, error)
	GetContractAddressesByNetworkAndLevel(string, int64) (gjson.Result, error)
	GetContracts(map[string]interface{}) ([]models.Contract, error)
	GetContractsIDByAddress([]string, string) ([]string, error)
	GetAffectedContracts(string, int64, int64) ([]string, error)
	IsFAContract(string, string) (bool, error)
	NeedParseOperation(string, string, string) (bool, error)
	RecalcContractStats(string, string) (ContractStats, error)
	UpdateContractMigrationsCount(string, string) error
	GetDAppStats(string, []string, string) (DAppStats, error)
	GetContractTransfers(string, string, int64, int64) (TransfersResponse, error)
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
	GetAllLevelsForNetwork(string) (map[int64]struct{}, error)
	GetOperations(map[string]interface{}, int64, bool) ([]models.Operation, error)
}

// IProjects -
type IProjects interface {
	GetProjectsLastContract() ([]models.Contract, error)
	GetSameContracts(models.Contract, int64, int64) (SameContractsResponse, error)
	GetSimilarContracts(models.Contract, int64, int64) ([]SimilarContract, uint64, error)
	GetProjectsStats() ([]ProjectStats, error)
	GetDiffTasks() ([]DiffTask, error)
}

// IProtocol -
type IProtocol interface {
	GetProtocol(string, string, int64) (models.Protocol, error)
	GetSymLinks(string, int64) (map[string]bool, error)
}

// ISearch -
type ISearch interface {
	SearchByText(string, int64, []string, map[string]interface{}, bool) (SearchResult, error)
}

// ISnapshot -
type ISnapshot interface {
	CreateAWSRepository(string, string, string) error
	ListRepositories() ([]string, error)
	CreateSnapshots(string, string, []string) error
	RestoreSnapshots(string, string, []string) error
	ListSnapshots(string) (string, error)
	SetSnapshotPolicy(string, string, string, string, int64) error
	GetAllPolicies() ([]string, error)
	GetMappings([]string) (map[string]string, error)
	CreateMapping(string, io.Reader) error
}

// IStats -
type IStats interface {
	GetNetworkCountStats(string) (NetworkCountStats, error)
	GetDateHistogram(string, string, string, string, string, []string) ([][]int64, error)
	GetCallsCountByNetwork() (map[string]int64, error)
	GetContractStatsByNetwork() (map[string]ContractCountStats, error)
	GetFACountByNetwork() (map[string]int64, error)
	GetLanguagesForNetwork(network string) (map[string]int64, error)
}

// ITokens -
type ITokens interface {
	GetTokens(string, string, int64, int64) ([]models.Contract, int64, error)
	GetTokenTransferOperations(string, string, string, int64) (PageableOperations, error)
	GetTokensStats(string, []string, []string) (map[string]TokenUsageStats, error)
	GetTokenVolumeSeries(string, string, []string, uint) ([][]int64, error)
}

// IElastic -
type IElastic interface {
	IGeneral
	IBigMap
	IBigMapDiff
	IBlock
	IBulk
	IContract
	IEvents
	IMigrations
	IOperations
	IProjects
	IProtocol
	ISearch
	ISnapshot
	IStats
	ITokens
}
