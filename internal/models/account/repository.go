package account

// Repository -
type Repository interface {
	Get(address string) (Account, error)
	Alias(address string) (string, error)
	UpdateAlias(account Account) error
}
