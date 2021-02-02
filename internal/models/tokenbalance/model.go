package tokenbalance

import (
	"fmt"
	"math/big"

	jsoniter "github.com/json-iterator/go"
	"github.com/sirupsen/logrus"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// TokenBalance -
type TokenBalance struct {
	Network  string `json:"network"`
	Address  string `json:"address"`
	Contract string `json:"contract"`
	TokenID  int64  `json:"token_id"`
	Balance  string `json:"balance"`

	Value *big.Int `json:"-"`
}

// GetID -
func (tb *TokenBalance) GetID() string {
	return fmt.Sprintf("%s_%s_%s_%d", tb.Network, tb.Address, tb.Contract, tb.TokenID)
}

// GetIndex -
func (tb *TokenBalance) GetIndex() string {
	return "token_balance"
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
		"balance":  tb.Balance,
	}
}

// Sum -
func (tb *TokenBalance) Sum(delta *TokenBalance) {
	tb.Value.Add(tb.Value, delta.Value)
}

// UnmarshalJSON -
func (tb *TokenBalance) UnmarshalJSON(data []byte) error {
	type buf TokenBalance
	if err := json.Unmarshal(data, (*buf)(tb)); err != nil {
		return err
	}
	tb.Value = big.NewInt(0)

	if _, ok := tb.Value.SetString(tb.Balance, 10); !ok {
		return fmt.Errorf("Can't set balance value: %s", tb.Balance)
	}
	return nil
}

// MarshalJSON -
func (tb *TokenBalance) MarshalJSON() ([]byte, error) {
	if tb.Value == nil {
		return nil, fmt.Errorf("Nil balance value")
	}
	tb.Balance = tb.Value.String()
	type buf TokenBalance
	return json.Marshal((*buf)(tb))
}
