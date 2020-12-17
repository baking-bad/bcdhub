package tezosdomain

// Repository -
type Repository interface {
	ListDomains(network string, size, offset int64) (DomainsResponse, error)
	ResolveDomainByAddress(network string, address string) (*TezosDomain, error)
}
