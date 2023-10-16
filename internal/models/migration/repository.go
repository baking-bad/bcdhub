package migration

import "context"

//go:generate mockgen -source=$GOFILE -destination=../mock/migration/mock.go -package=migration -typed
type Repository interface {
	Get(ctx context.Context, contractID int64) ([]Migration, error)
}
