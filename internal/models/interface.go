package models

import "github.com/baking-bad/bcdhub/internal/models/types"

// GeneralRepository -
type GeneralRepository interface {
	CreateIndexes() error
	DeleteIndices(indices []string) error
	DeleteByContract(network types.Network, indices []string, address string) error
	GetByID(output Model) error
	GetByNetwork(network types.Network, index string) ([]Model, error)
	UpdateDoc(model Model) (err error)
	UpdateFields(index string, id int64, data interface{}, fields ...string) error

	GetNetworkCountStats(network types.Network) (map[string]int64, error)
	GetDateHistogram(period string, opts ...HistogramOption) ([][]float64, error)

	// GetStats - returns full stats for network(s). If `network` is not empty returns stats only for that network.
	GetStats(network types.Network) (map[string]*NetworkStats, error)

	GetLanguagesForNetwork(network types.Network) (map[string]int64, error)
	IsRecordNotFound(err error) bool

	// Save - performs insert or update items.
	Save(items []Model) error
	BulkDelete([]Model) error
}
