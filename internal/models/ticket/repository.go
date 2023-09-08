package ticket

//go:generate mockgen -source=$GOFILE -destination=../mock/ticket/mock.go -package=ticket -typed
type Repository interface {
	Get(ticketer string, limit, offset int64) ([]TicketUpdate, error)
	Has(contractID int64) (bool, error)
	ForOperation(operationId int64) ([]TicketUpdate, error)
}
