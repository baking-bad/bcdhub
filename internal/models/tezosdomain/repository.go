package tezosdomain

import "github.com/baking-bad/bcdhub/internal/models/types"

// Repository -
type Repository interface {
	ListDomains(network types.Network, size, offset int64) (DomainsResponse, error)
	ResolveDomainByAddress(network types.Network, address string) (*TezosDomain, error)
}
