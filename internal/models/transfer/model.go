package transfer

import (
	"fmt"
	"time"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
	"github.com/shopspring/decimal"
)

// Transfer -
type Transfer struct {
	// nolint
	tableName struct{} `pg:"transfers"`

	ID       int64
	Network  types.Network `pg:",type:SMALLINT"`
	Contract string

	InitiatorID int64
	Initiator   account.Account `pg:",rel:has-one"`

	FromID int64
	From   account.Account `pg:",rel:has-one"`

	ToID int64
	To   account.Account `pg:",rel:has-one"`

	Status     types.OperationStatus `pg:",type:SMALLINT"`
	Timestamp  time.Time
	Level      int64            `pg:",use_zero"`
	TokenID    uint64           `pg:",type:numeric(50,0),use_zero"`
	Amount     decimal.Decimal  `pg:",type:numeric(100,0),use_zero"`
	Parent     types.NullString `pg:",type:text"`
	Entrypoint string

	OperationID int64
}

// GetID -
func (t *Transfer) GetID() int64 {
	return t.ID
}

// GetIndex -
func (t *Transfer) GetIndex() string {
	return "transfers"
}

// Save -
func (t *Transfer) Save(tx pg.DBI) error {
	_, err := tx.Model(t).Returning("id").Insert()
	return err
}

// LogFields -
func (t *Transfer) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network":  t.Network.String(),
		"contract": t.Contract,
		"block":    t.Level,
		"from":     t.From,
		"to":       t.To,
	}
}

// GetFromTokenBalanceID -
func (t *Transfer) GetFromTokenBalanceID() string {
	if t.From.Address != "" {
		return fmt.Sprintf("%s_%s_%d", t.From.Address, t.Contract, t.TokenID)
	}
	return ""
}

// GetToTokenBalanceID -
func (t *Transfer) GetToTokenBalanceID() string {
	if t.To.Address != "" {
		return fmt.Sprintf("%s_%s_%d", t.To.Address, t.Contract, t.TokenID)
	}
	return ""
}

// MakeTokenBalanceUpdate -
func (t *Transfer) MakeTokenBalanceUpdate(from, rollback bool) *tokenbalance.TokenBalance {
	tb := &tokenbalance.TokenBalance{
		Network:  t.Network,
		Contract: t.Contract,
		TokenID:  t.TokenID,
		Balance:  decimal.Zero,
	}
	switch {
	case from && rollback:
		tb.Account = t.From
		tb.AccountID = t.FromID
		tb.Balance = t.Amount
	case !from && rollback:
		tb.Account = t.To
		tb.AccountID = t.ToID
		tb.Balance = t.Amount.Neg()
	case from && !rollback:
		tb.Account = t.From
		tb.AccountID = t.FromID
		tb.Balance = t.Amount.Neg()
	case !from && !rollback:
		tb.Account = t.To
		tb.AccountID = t.ToID
		tb.Balance = t.Amount
	}
	return tb
}

// Pageable -
type Pageable struct {
	Transfers []Transfer `json:"transfers"`
	Total     int64      `json:"total"`
	LastID    string     `json:"last_id"`
}

// Balance -
type Balance struct {
	Balance decimal.Decimal
	Address string
	TokenID uint64
}
