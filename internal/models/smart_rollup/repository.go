package smartrollup

//go:generate mockgen -source=$GOFILE -destination=../mock/smart_rollup/mock.go -package=smart_rollup -typed
type Repository interface {
	Get(address string) (SmartRollup, error)
	List(limit, offset int64, sort string) ([]SmartRollup, error)
}
