package transfer

import (
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
)

// BalanceParser -
type BalanceParser interface {
	Parse(balances []events.TokenBalance) ([]*transfer.Transfer, error)
}

// DefaultBalanceParser -
type DefaultBalanceParser struct{}

// NewDefaultBalanceParser -
func NewDefaultBalanceParser() *DefaultBalanceParser {
	return new(DefaultBalanceParser)
}

// Parse -
func (parser *DefaultBalanceParser) Parse(balances []events.TokenBalance, operation operation.Operation) ([]*transfer.Transfer, error) {
	transfers := make([]*transfer.Transfer, 0)
	for _, balance := range balances {
		transfer := transfer.EmptyTransfer(operation)
		if balance.Value > 0 {
			transfer.To = balance.Address
		} else {
			transfer.From = balance.Address
		}
		transfer.Amount = float64(balance.Value)
		transfer.TokenID = balance.TokenID

		transfers = append(transfers, transfer)
	}

	return transfers, nil
}
