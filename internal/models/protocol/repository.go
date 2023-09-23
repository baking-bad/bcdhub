package protocol

import "context"

//go:generate mockgen -source=$GOFILE -destination=../mock/protocol/mock.go -package=protocol -typed
type Repository interface {
	Get(ctx context.Context, hash string, level int64) (Protocol, error)
	GetAll(ctx context.Context) (response []Protocol, err error)
	GetByNetworkWithSort(ctx context.Context, sortField, order string) (response []Protocol, err error)
	GetByID(ctx context.Context, id int64) (response Protocol, err error)
}
