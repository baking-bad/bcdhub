package smartrollup

import "context"

//go:generate mockgen -source=$GOFILE -destination=../mock/smart_rollup/mock.go -package=smart_rollup -typed
type Repository interface {
	Get(ctx context.Context, address string) (SmartRollup, error)
	List(ctx context.Context, limit, offset int64, sort string) ([]SmartRollup, error)
}
