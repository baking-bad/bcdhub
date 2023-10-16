package operation

import (
	"context"
)

//go:generate mockgen -source=$GOFILE -destination=../mock/operation/mock.go -package=operation -typed
type Repository interface {
	Last(ctx context.Context, filter map[string]interface{}, lastID int64) (Operation, error)
	GetByHash(ctx context.Context, hash []byte) ([]Operation, error)
	GetByHashAndCounter(ctx context.Context, hash []byte, counter int64) ([]Operation, error)
	OPG(ctx context.Context, address string, size, lastID int64) ([]OPG, error)
	Origination(ctx context.Context, accountID int64) (Operation, error)
	GetByID(ctx context.Context, id int64) (Operation, error)
	ListEvents(ctx context.Context, accountID int64, size, offset int64) ([]Operation, error)
}
