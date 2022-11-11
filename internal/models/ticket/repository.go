package ticket

// Repository -
type Repository interface {
	Get(contract string, limit, offset int) ([]TicketUpdate, error)
}
