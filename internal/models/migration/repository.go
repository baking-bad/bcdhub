package migration

// Repository -
type Repository interface {
	GetMigrations(string, string) ([]Migration, error)
}
