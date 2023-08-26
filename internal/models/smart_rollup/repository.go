package smartrollup

// Repository -
type Repository interface {
	Get(address string) (SmartRollup, error)
	List(limit, offset int64, sort string) ([]SmartRollup, error)
}
