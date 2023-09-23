package account

import "context"

//go:generate mockgen -source=$GOFILE -destination=../mock/account/mock.go -package=account -typed
type Repository interface {
	Get(ctx context.Context, address string) (Account, error)
}
