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

	ID          int64     `bun:"id,pk,notnull,autoincrement"`
	Timestamp   time.Time `bun:"timestamp,pk,notnull"`
	OperationID int64
	Level       int64
	TicketerID  int64
	Ticketer    account.Account `bun:"rel:belongs-to"`
	ContentType []byte
	Content     []byte
	AccountID   int64
	Account     account.Account `bun:"rel:belongs-to"`
	Amount      decimal.Decimal `bun:"amount,type:numeric(200,0)"`
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
		"id":          t.ID,
		"block":       t.Level,
		"ticketer_id": t.TicketerID,
	}
}
