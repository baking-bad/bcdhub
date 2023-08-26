package account

// Repository -
type Repository interface {
	Get(address string) (Account, error)
}
