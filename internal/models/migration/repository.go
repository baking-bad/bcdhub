package migration

// Repository -
type Repository interface {
	Get(string, string) ([]Migration, error)
	Count(string, string) (int64, error)
	GetByIDs(ids ...int64) ([]Migration, error)
}
