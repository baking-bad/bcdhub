package balanceupdate

// Repository -
type Repository interface {
	GetBalance(network, address string) (int64, error)
}
