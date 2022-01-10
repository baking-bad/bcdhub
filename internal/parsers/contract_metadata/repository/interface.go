package repository

// Repository -
type Repository interface {
	GetAll() ([]Item, error)
	Get(network, name string) (Item, error)
}
