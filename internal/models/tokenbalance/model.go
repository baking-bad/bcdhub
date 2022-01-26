package tokenbalance

import (
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TokenBalance -
type TokenBalance struct {
	ID       int64           `json:"-" gorm:"autoIncrement:true;not null;"`
	Network  types.Network   `json:"network" gorm:"type:SMALLINT;not null;primaryKey;index:token_balances_token_idx;default:0"`
	Address  string          `json:"address" gorm:"not null;primaryKey"`
	Contract string          `json:"contract" gorm:"not null;primaryKey;index:token_balances_token_idx"`
	TokenID  uint64          `json:"token_id" gorm:"type:numeric(50,0);default:0;primaryKey;autoIncrement:false;index:token_balances_token_idx"`
	Balance  decimal.Decimal `json:"balance" gorm:"type:numeric(200,0);default:0"`

	IsLedger bool `json:"-" gorm:"-"`
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
func (tb *TokenBalance) Save(tx *gorm.DB) error {
	var s clause.Set

	if tb.IsLedger {
		s = clause.Assignments(map[string]interface{}{
			"balance": tb.Balance,
		})
	} else {
		s = clause.Assignments(map[string]interface{}{
			"balance": gorm.Expr("token_balances.balance + ?", tb.Balance),
		})
	}

	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "network"},
			{Name: "contract"},
			{Name: "address"},
			{Name: "token_id"},
		},
		DoUpdates: s,
	}).Create(tb).Error
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
