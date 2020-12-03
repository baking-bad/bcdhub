package transfer

import (
	"github.com/baking-bad/bcdhub/internal/events"
	"github.com/baking-bad/bcdhub/internal/models"
)

// BalanceParser -
type BalanceParser interface {
	Parse(balances []events.TokenBalance) ([]*models.Transfer, error)
}

// DefaultBalanceParser -
type DefaultBalanceParser struct{}

// NewDefaultBalanceParser -
func NewDefaultBalanceParser() *DefaultBalanceParser {
	return new(DefaultBalanceParser)
}

// Parse -
func (parser *DefaultBalanceParser) Parse(balances []events.TokenBalance, operation models.Operation) ([]*models.Transfer, error) {
	transfers := make([]*models.Transfer, 0)
	for _, balance := range balances {
		transfer := models.EmptyTransfer(operation)
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
