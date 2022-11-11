package ticket

// Repository -
type Repository interface {
	Get(ticketer string, limit, offset int64) ([]TicketUpdate, error)
}
