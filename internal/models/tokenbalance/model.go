package tokenbalance

import (
	"fmt"
	"math/big"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// TokenBalance -
type TokenBalance struct {
	ID       int64  `json:"-" gorm:"autoIncrement:true"`
	Network  string `json:"network" gorm:"not null;primaryKey"`
	Address  string `json:"address" gorm:"not null;primaryKey"`
	Contract string `json:"contract" gorm:"not null;primaryKey"`
	TokenID  uint64 `json:"token_id" gorm:"type:numeric(50,0);default:0;primaryKey;autoIncrement:false"`
	Balance  string `json:"balance" gorm:"balance,default:0"`

	Value *big.Int `json:"-" gorm:"-"`
}

// GetID -
func (tb *TokenBalance) GetID() int64 {
	return tb.ID
}

// GetIndex -
func (tb *TokenBalance) GetIndex() string {
	return "token_balances"
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
		"balance":  tb.Value.String(),
	}
}

// Sum -
func (tb *TokenBalance) Sum(delta *TokenBalance) {
	tb.Value.Add(tb.Value, delta.Value)
}

func (tb *TokenBalance) unmarshal() error {
	tb.Value = big.NewInt(0)

	if _, ok := tb.Value.SetString(tb.Balance, 10); !ok {
		return fmt.Errorf("Can't set balance value: %s", tb.Balance)
	}
	return nil
}

func (tb *TokenBalance) marshal() error {
	if tb.Value == nil {
		return fmt.Errorf("Nil balance value")
	}
	tb.Balance = tb.Value.String()
	return nil
}

// UnmarshalJSON -
func (tb *TokenBalance) UnmarshalJSON(data []byte) error {
	type buf TokenBalance
	if err := json.Unmarshal(data, (*buf)(tb)); err != nil {
		return err
	}
	return tb.unmarshal()
}

// MarshalJSON -
func (tb *TokenBalance) MarshalJSON() ([]byte, error) {
	if err := tb.marshal(); err != nil {
		return nil, err
	}
	type buf TokenBalance
	return json.Marshal((*buf)(tb))
}

// BeforeScan -
func (tb *TokenBalance) AfterFind(tx *gorm.DB) error {
	return tb.unmarshal()
}

// BeforeInsert -
func (tb *TokenBalance) BeforeCreate(tx *gorm.DB) error {
	return tb.marshal()
}

// BeforeUpdate -
func (tb *TokenBalance) BeforeUpdate(tx *gorm.DB) error {
	return tb.marshal()
}
