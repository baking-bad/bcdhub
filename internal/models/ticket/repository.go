package ticket

import "context"

//go:generate mockgen -source=$GOFILE -destination=../mock/ticket/mock.go -package=ticket -typed
type Repository interface {
	Updates(ctx context.Context, ticketer string, limit, offset int64) ([]TicketUpdate, error)
	UpdatesForOperation(ctx context.Context, operationId int64) ([]TicketUpdate, error)
}
