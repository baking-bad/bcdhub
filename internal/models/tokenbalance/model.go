package tokenbalance

import (
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/go-pg/pg/v10"
	"github.com/shopspring/decimal"
)

// TokenBalance -
type TokenBalance struct {
	// nolint
	tableName struct{} `pg:"token_balances"`

	ID       int64           `pg:",notnull"`
	Network  types.Network   `pg:",type:SMALLINT,notnull,unique:token_balance,use_zero"`
	Address  string          `pg:",notnull,unique:token_balance"`
	Contract string          `pg:",notnull,unique:token_balance"`
	TokenID  uint64          `pg:",type:numeric(50,0),use_zero,unique:token_balance"`
	Balance  decimal.Decimal `pg:",type:numeric(100,0),use_zero"`

	IsLedger bool `pg:"-"`
}

// GetID -
func (tb *TokenBalance) GetID() int64 {
	return tb.ID
}

// GetIndex -
func (tb *TokenBalance) GetIndex() string {
	return "token_balances"
}

// Constraint -
func (tb *TokenBalance) Save(tx pg.DBI) error {
	query := tx.Model(tb).OnConflict("(network, contract, address, token_id) DO UPDATE")

	if tb.IsLedger {
		query.Set("balance = excluded.balance")
	} else {
		query.Set("balance = token_balance.balance + excluded.balance")
	}

	_, err := query.Returning("id").Insert()
	return err
}

// LogFields -
func (tb *TokenBalance) LogFields() map[string]interface{} {
	return map[string]interface{}{
		"network":   tb.Network.String(),
		"address":   tb.Address,
		"contract":  tb.Contract,
		"token_id":  tb.TokenID,
		"balance":   tb.Balance.String(),
		"is_ledger": tb.IsLedger,
	}
}
