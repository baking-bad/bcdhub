package account

import (
	"context"
)

//go:generate mockgen -source=$GOFILE -destination=../mock/account/mock.go -package=account -typed
type Repository interface {
	Get(ctx context.Context, address string) (Account, error)
	RecentlyCalledContracts(ctx context.Context, offset, size int64) (accounts []Account, err error)
}
