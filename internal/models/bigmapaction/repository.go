package bigmapaction

import "context"

//go:generate mockgen -source=$GOFILE -destination=../mock/bigmapaction/mock.go -package=bigmapaction -typed
type Repository interface {
	Get(ctx context.Context, ptr, limit, offset int64) ([]BigMapAction, error)
}
