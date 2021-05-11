package models

// GeneralRepository -
type GeneralRepository interface {
	CreateIndexes() error
	DeleteIndices(indices []string) error
	DeleteByContract(indices []string, network, address string) error
	GetByID(output Model) error
	GetByNetwork(network, index string) ([]Model, error)
	UpdateDoc(model Model) (err error)
	UpdateFields(index string, id int64, data interface{}, fields ...string) error
	GetEvents([]SubscriptionRequest, int64, int64) ([]Event, error)

	GetNetworkCountStats(string) (map[string]int64, error)
	GetDateHistogram(period string, opts ...HistogramOption) ([][]float64, error)

	// GetStats - returns full stats for network(s). If `network` is not empty returns stats only for that network.
	GetStats(network string) (map[string]*NetworkStats, error)

	GetLanguagesForNetwork(network string) (map[string]int64, error)
	IsRecordNotFound(err error) bool

	// Save - performs insert or update items.
	Save(items []Model) error
	BulkDelete([]Model) error
	SetAlias(network, address, alias string) error
}
