package bigmapdiff

import "context"

//go:generate mockgen -source=$GOFILE -destination=../mock/bigmapdiff/mock.go -package=bigmapdiff -typed
type Repository interface {
	Get(ctx context.Context, reqCtx GetContext) ([]Bucket, error)
	GetByAddress(ctx context.Context, address string) ([]BigMapDiff, error)
	GetForOperation(ctx context.Context, id int64) ([]BigMapDiff, error)
	GetByPtr(ctx context.Context, contract string, ptr int64) ([]BigMapState, error)
	GetByPtrAndKeyHash(ctx context.Context, ptr int64, keyHash string, size int64, offset int64) ([]BigMapDiff, int64, error)
	GetForAddress(ctx context.Context, address string) ([]BigMapState, error)
	Count(ctx context.Context, ptr int64) (int, error)
	Current(ctx context.Context, keyHash string, ptr int64) (BigMapState, error)
	Previous(ctx context.Context, diffs []BigMapDiff) ([]BigMapDiff, error)
	GetStats(ctx context.Context, ptr int64) (Stats, error)
	Keys(ctx context.Context, reqCtx GetContext) (states []BigMapState, err error)
}
