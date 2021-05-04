package tokenbalance

import (
	"math/big"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// TokenBalance -
type TokenBalance struct {
	ID            int64   `json:"-" gorm:"autoIncrement:true"`
	Network       string  `json:"network" gorm:"not null;primaryKey;index:token_balances_token_idx"`
	Address       string  `json:"address" gorm:"not null;primaryKey"`
	Contract      string  `json:"contract" gorm:"not null;primaryKey;index:token_balances_token_idx"`
	TokenID       uint64  `json:"token_id" gorm:"type:numeric(50,0);default:0;primaryKey;autoIncrement:false;index:token_balances_token_idx"`
	Balance       float64 `json:"balance" gorm:"type:numeric(100,0);default:0"`
	BalanceString string  `json:"balance_string"`

	IsLedger bool     `json:"-" gorm:"-"`
	Value    *big.Int `json:"-" gorm:"-"`
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

	f, _ := new(big.Float).SetInt(tb.Value).Float64()
	if tb.IsLedger {
		s = clause.Assignments(map[string]interface{}{
			"balance":        f,
			"balance_string": tb.Value.String(),
		})
	} else {
		s = clause.Assignments(map[string]interface{}{
			"balance":        gorm.Expr("token_balances.balance + ?", f),
			"balance_string": gorm.Expr("(token_balances.balance + ?)::text", f),
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
	}).Save(tb).Error
}

// GetQueues -
func (tb *TokenBalance) GetQueues() []string {
	return nil
}

// MarshalToQueue -
func (tb *TokenBalance) MarshalToQueue() ([]byte, error) {
	return nil, nil
}

// LogFields -
func (tb *TokenBalance) LogFields() logrus.Fields {
	return logrus.Fields{
		"network":  tb.Network,
		"address":  tb.Address,
		"contract": tb.Contract,
		"token_id": tb.TokenID,
	}
}

// AfterFind -
func (tb *TokenBalance) AfterFind(tx *gorm.DB) (err error) {
	return tb.unmarshal()
}

// BeforeCreate -
func (tb *TokenBalance) BeforeCreate(tx *gorm.DB) (err error) {
	return tb.marshal()
}

// BeforeUpdate -
func (tb *TokenBalance) BeforeUpdate(tx *gorm.DB) (err error) {
	return tb.marshal()
}

func (tb *TokenBalance) marshal() error {
	if tb.Value == nil {
		return errors.New("Nil amount in transfer")
	}
	tb.BalanceString = tb.Value.String()
	tb.Balance, _ = new(big.Float).SetInt(tb.Value).Float64()
	return nil
}

func (tb *TokenBalance) unmarshal() error {
	if tb.Value == nil {
		tb.Value = big.NewInt(0)
	}

	if _, ok := tb.Value.SetString(tb.BalanceString, 10); !ok {
		return errors.Errorf("Invalid amount in transfer: %s", tb.BalanceString)
	}

	return nil
}
