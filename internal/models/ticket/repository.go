package ticket

import "context"

type BalanceRequest struct {
	Limit               int64
	Offset              int64
	WithoutZeroBalances bool
}

type UpdatesRequest struct {
	Account  string
	Ticketer string
	TicketId *uint64
	Limit    int64
	Offset   int64
}

//go:generate mockgen -source=$GOFILE -destination=../mock/ticket/mock.go -package=ticket -typed
type Repository interface {
	List(ctx context.Context, ticketer string, limit, offset int64) ([]Ticket, error)
	Updates(ctx context.Context, req UpdatesRequest) ([]TicketUpdate, error)
	UpdatesForOperation(ctx context.Context, operationId int64) ([]TicketUpdate, error)
	BalancesForAccount(ctx context.Context, accountId int64, req BalanceRequest) ([]Balance, error)
}
