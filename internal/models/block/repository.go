package block

import "context"

//go:generate mockgen -source=$GOFILE -destination=../mock/block/mock.go -package=block -typed
type Repository interface {
	Get(ctx context.Context, level int64) (Block, error)
	Last(ctx context.Context) (Block, error)
}
