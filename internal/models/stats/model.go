package stats

import (
	"github.com/uptrace/bun"
)

// Stats - entity for blockchain general stats
type Stats struct {
	bun.BaseModel `bun:"stats"`

	ID                          int64 `bun:"id,pk,notnull,autoincrement"`
	ContractsCount              int   `bun:"contracts_count"`
	SmartRollupsCount           int   `bun:"smart_rollups_count"`
	GlobalConstantsCount        int   `bun:"global_constants_count"`
	OperationsCount             int   `bun:"operations_count"`
	EventsCount                 int   `bun:"events_count"`
	TransactionsCount           int   `bun:"tx_count"`
	OriginationsCount           int   `bun:"originations_count"`
	SrOriginationsCount         int   `bun:"sr_originations_count"`
	SrExecutesCount             int   `bun:"sr_executes_count"`
	RegisterGlobalConstantCount int   `bun:"register_global_constants_count"`
	TransferTicketsCount        int   `bun:"transfer_tickets_count"`
}

// GetID -
func (sr *Stats) GetID() int64 {
	return sr.ID
}

func (Stats) TableName() string {
	return "stats"
}
