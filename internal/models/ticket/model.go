package ticket

import (
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/go-pg/pg/v10"
	"github.com/shopspring/decimal"
)

// TicketUpdate -
type TicketUpdate struct {
	// nolint
	tableName struct{} `pg:"ticket_updates"`

	ID          int64
	OperationID int64
	Level       int64 `pg:",use_zero"`
	Timestamp   time.Time
	TicketerID  int64
	Ticketer    account.Account `pg:",rel:has-one"`
	ContentType []byte
	Content     []byte
	AccountID   int64
	Account     account.Account `pg:",rel:has-one"`
	Amount      decimal.Decimal `pg:",type:numeric(200,0),use_zero"`
}

// GetID -
func (t *TicketUpdate) GetID() int64 {
	return t.ID
}

// GetIndex -
func (t *TicketUpdate) GetIndex() string {
	return "ticket_updates"
}

// Save -
func (t *TicketUpdate) Save(tx pg.DBI) error {
	_, err := tx.Model(t).Returning("id").Insert()
	return err
}

// LogFields -
func (t *TicketUpdate) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"id":          t.ID,
		"block":       t.Level,
		"ticketer_id": t.TicketerID,
	}
}
