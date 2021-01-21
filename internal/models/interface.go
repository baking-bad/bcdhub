package models

import (
	"io"
)

// GeneralRepository -
type GeneralRepository interface {
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
	GetEvents([]SubscriptionRequest, int64, int64) ([]Event, error)
	SearchByText(string, int64, []string, map[string]interface{}, bool) (Result, error)
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
	GetNetworkCountStats(string) (map[string]int64, error)
	GetDateHistogram(period string, opts ...HistogramOption) ([][]int64, error)
	// GetCallsCountByNetwork - returns contract calls splitted by network. If `network` is not empty returns stats only for that network.
	GetCallsCountByNetwork(network string) (map[string]int64, error)
	// GetContractStatsByNetwork - returns contract stats splitted by network. If `network` is not empty returns stats only for that network.
	GetContractStatsByNetwork(network string) (map[string]ContractCountStats, error)
	// GetFACountByNetwork - returns FA contracts count splitted by network. If `network` is not empty returns stats only for that network.
	GetFACountByNetwork(network string) (map[string]int64, error)
	GetLanguagesForNetwork(network string) (map[string]int64, error)
	IsRecordNotFound(err error) bool
	BulkInsert([]Model) error
	BulkUpdate([]Model) error
	BulkDelete([]Model) error
	BulkRemoveField(string, []Model) error
	SetAlias(network, address, alias string) error
}
