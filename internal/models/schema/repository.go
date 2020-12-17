package schema

// Repository -
type Repository interface {
	Get(address string) (Schema, error)
}
