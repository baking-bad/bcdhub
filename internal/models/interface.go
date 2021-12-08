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

	IsRecordNotFound(err error) bool

	// Save - performs insert or update items.
	Save(items []Model) error
	BulkDelete([]Model) error
}

// Statistics -
type Statistics interface {
	NetworkCountStats(network types.Network) (map[string]int64, error)
	Histogram(period string, opts ...HistogramOption) ([][]float64, error)
	// GetStats - returns full stats for network(s). If `network` is not empty returns stats only for that network.
	NetworkStats(network types.Network) (map[string]*NetworkStats, error)
	LanguageByNetwork(network types.Network) (map[string]int64, error)
}
