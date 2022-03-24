package global_constant

// Repository -
type Repository interface {
	Get(address string) (GlobalConstant, error)
	All(addresses ...string) ([]GlobalConstant, error)
}
