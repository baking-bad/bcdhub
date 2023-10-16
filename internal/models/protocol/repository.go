package protocol

import "context"

//go:generate mockgen -source=$GOFILE -destination=../mock/protocol/mock.go -package=protocol -typed
type Repository interface {
	Get(ctx context.Context, hash string, level int64) (Protocol, error)
	GetByID(ctx context.Context, id int64) (response Protocol, err error)
}
