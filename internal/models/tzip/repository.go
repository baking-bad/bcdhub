package tzip

// Repository -
type Repository interface {
	Get(network, address string) (*TZIP, error)
	GetWithEvents() ([]TZIP, error)
	GetLastIDWithEvents() (int64, error)
	GetBySlug(slug string) (*TZIP, error)
	GetAliases(network string) ([]TZIP, error)
	GetAliasesMap(network string) (map[string]string, error)
}
