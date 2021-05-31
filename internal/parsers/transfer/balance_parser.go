package transfer

import (
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	tbParser "github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/shopspring/decimal"
)

// DefaultBalanceParser -
type DefaultBalanceParser struct {
	repo tokenbalance.Repository
}

// NewDefaultBalanceParser -
func NewDefaultBalanceParser(repo tokenbalance.Repository) *DefaultBalanceParser {
	return &DefaultBalanceParser{repo}
}

// Parse -
func (parser *DefaultBalanceParser) Parse(balances []tbParser.TokenBalance, operation operation.Operation) ([]*transfer.Transfer, error) {
	transfers := make([]*transfer.Transfer, 0)
	for _, balance := range balances {
		transfer := transfer.EmptyTransfer(operation)
		if balance.Value.Cmp(decimal.Zero) > 0 {
			transfer.To = balance.Address
		} else {
			transfer.From = balance.Address
		}
		transfer.Amount = balance.Value.Abs()
		transfer.TokenID = balance.TokenID

		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// ParseBalances -
func (parser *DefaultBalanceParser) ParseBalances(network types.Network, contract string, balances []tbParser.TokenBalance, operation operation.Operation) ([]*transfer.Transfer, error) {
	transfers := make([]*transfer.Transfer, 0)
	for _, balance := range balances {
		transfer := transfer.EmptyTransfer(operation)

		tb, err := parser.repo.Get(network, contract, balance.Address, balance.TokenID)
		if err != nil {
			return nil, err
		}

		delta := balance.Value.Sub(tb.Balance)
		if delta.Cmp(decimal.Zero) > 0 {
			transfer.To = balance.Address
		} else {
			transfer.From = balance.Address
		}

		transfer.Amount = delta.Abs()
		transfer.TokenID = balance.TokenID

		transfers = append(transfers, transfer)
	}

	return transfers, nil
}
