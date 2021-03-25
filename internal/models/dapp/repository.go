package dapp

// Repository -
type Repository interface {
	Get(slug string) (DApp, error)
	All() ([]DApp, error)
}
