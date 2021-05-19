package tzip

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, address string) (*TZIP, error)
	GetWithEvents() ([]TZIP, error)
	GetLastIDWithEvents() (int64, error)
	GetBySlug(slug string) (*TZIP, error)
	GetAliases(network types.Network) ([]TZIP, error)
	GetAliasesMap(network types.Network) (map[string]string, error)
}
