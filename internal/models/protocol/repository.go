package protocol

// Repository -
type Repository interface {
	Get(hash string, level int64) (Protocol, error)
	GetAll() (response []Protocol, err error)
	GetByNetworkWithSort(sortField, order string) (response []Protocol, err error)
	GetByID(id int64) (response Protocol, err error)
}
