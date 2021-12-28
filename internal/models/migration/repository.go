package migration

// Repository -
type Repository interface {
	Get(contractID int64) ([]Migration, error)
}
