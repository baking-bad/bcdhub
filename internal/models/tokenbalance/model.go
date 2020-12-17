package tokenbalance

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// TokenBalance -
type TokenBalance struct {
	Network  string `json:"network"`
	Address  string `json:"address"`
	Contract string `json:"contract"`
	TokenID  int64  `json:"token_id"`
	Balance  int64  `json:"balance"`
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
