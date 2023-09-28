package account

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/uptrace/bun"
)

// Account -
type Account struct {
	bun.BaseModel `bun:"accounts"`

	ID                 int64             `bun:"id,pk,notnull,autoincrement"`
	Type               types.AccountType `bun:"type,type:SMALLINT"`
	Address            string            `bun:"address,type:text,unique:address_hash"`
	Level              int64             `bun:"level"`
	LastAction         time.Time         `bun:"last_action"`
	OperationsCount    int64             `bun:"operations_count"`
	MigrationsCount    int64             `bun:"migrations_count"`
	EventsCount        int64             `bun:"events_count"`
	TicketUpdatesCount int64             `bun:"ticket_updates_count"`
}

// GetID -
func (a *Account) GetID() int64 {
	return a.ID
}

func (Account) TableName() string {
	return "accounts"
}

// IsEmpty -
func (a *Account) IsEmpty() bool {
	return a.Address == "" || a.Type == types.AccountTypeUnknown
}
