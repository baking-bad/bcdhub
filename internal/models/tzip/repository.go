package tzip

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	Get(network types.Network, address string) (*TZIP, error)
	GetWithEvents(updatedAt uint64) ([]TZIP, error)
	GetBySlug(slug string) (*TZIP, error)
	GetAliases(network types.Network) ([]TZIP, error)
}
