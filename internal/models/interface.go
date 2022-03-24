package models

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/models/types"
)

// GeneralRepository -
type GeneralRepository interface {
	CreateTables() error
	DeleteTables(indices []string) error
	DeleteByContract(indices []string, address string) error
	GetByID(output Model) error
	GetAll(index string) ([]Model, error)
	UpdateDoc(model Model) (err error)
	IsRecordNotFound(err error) bool

	// Save - performs insert or update items.
	Save(ctx context.Context, items []Model) error
	BulkDelete(context.Context, []Model) error
}

// Statistics -
type Statistics interface {
	NetworkCountStats(network types.Network) (map[string]int64, error)
	Histogram(period string, opts ...HistogramOption) ([][]float64, error)
	// NetworkStats - returns full stats for network(s). If `network` is not empty returns stats only for that network.
	NetworkStats(network types.Network) (map[string]*NetworkStats, error)
	// ContractStats - returns operations count and last action time
	ContractStats(network types.Network, address string) (ContractStats, error)
}
