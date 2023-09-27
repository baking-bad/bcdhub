package ticket

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

// TicketUpdate -
type TicketUpdate struct {
	bun.BaseModel `bun:"ticket_updates"`

	ID          int64           `bun:"id,pk,notnull,autoincrement"`
	Timestamp   time.Time       `bun:"timestamp,pk,notnull"`
	OperationID int64           `bun:"operation_id"`
	Level       int64           `bun:"level"`
	TicketId    int64           `bun:"ticket_id"`
	AccountID   int64           `bun:"account_id"`
	Amount      decimal.Decimal `bun:"amount,type:numeric(200,0)"`

	Account account.Account `bun:"rel:belongs-to"`
	Ticket  Ticket          `bun:"rel:belongs-to"`
}

// GetID -
func (t *TicketUpdate) GetID() int64 {
	return t.ID
}

func (TicketUpdate) TableName() string {
	return "ticket_updates"
}

// LogFields -
func (t *TicketUpdate) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"id":        t.ID,
		"block":     t.Level,
		"ticket_id": t.TicketId,
	}
}
