package stats

import "context"

//go:generate mockgen -source=$GOFILE -destination=../mock/stats/mock.go -package=stats -typed
type Repository interface {
	Get(ctx context.Context) (Stats, error)
}
