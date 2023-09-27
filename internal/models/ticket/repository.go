package ticket

import "context"

//go:generate mockgen -source=$GOFILE -destination=../mock/ticket/mock.go -package=ticket -typed
type Repository interface {
	Get(ctx context.Context, ticketer string, limit, offset int64) ([]TicketUpdate, error)
	ForOperation(ctx context.Context, operationId int64) ([]TicketUpdate, error)
}
