package ticket

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/shopspring/decimal"
	"github.com/uptrace/bun"
)

type Balance struct {
	bun.BaseModel `bun:"ticket_balances"`

	TicketId  int64           `bun:"ticket_id,pk,notnull"`
	AccountId int64           `bun:"account_id,pk,notnull"`
	Amount    decimal.Decimal `bun:"amount,type:numeric(200,0)"`

	Ticket  Ticket          `bun:"rel:belongs-to"`
	Account account.Account `bun:"rel:belongs-to"`
}

func (Balance) GetID() int64 {
	return 0
}

func (Balance) TableName() string {
	return "ticket_balances"
}

// LogFields -
func (b Balance) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"ticket_id":  b.TicketId,
		"account_id": b.AccountId,
		"amount":     b.Amount.String(),
	}
}

func (b Balance) String() string {
	return fmt.Sprintf("%d_%d", b.TicketId, b.AccountId)
}
