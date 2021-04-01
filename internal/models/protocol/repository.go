package protocol

// Repository -
type Repository interface {
	Get(string, string, int64) (Protocol, error)
	GetAll() (response []Protocol, err error)
	GetByNetworkWithSort(network, sortField, order string) (response []Protocol, err error)
}
