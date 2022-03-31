package transfer

import (
	"errors"

	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	"github.com/baking-bad/bcdhub/internal/models/types"
	tbParser "github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
	"github.com/go-pg/pg/v10"
	"github.com/shopspring/decimal"
)

// DefaultBalanceParser -
type DefaultBalanceParser struct {
	repo     tokenbalance.Repository
	accounts account.Repository
}

// NewDefaultBalanceParser -
func NewDefaultBalanceParser(repo tokenbalance.Repository, accounts account.Repository) *DefaultBalanceParser {
	return &DefaultBalanceParser{repo, accounts}
}

// Parse -
func (parser *DefaultBalanceParser) Parse(balances []tbParser.TokenBalance, operation operation.Operation) ([]*transfer.Transfer, error) {
	transfers := make([]*transfer.Transfer, 0)
	for _, balance := range balances {
		transfer := operation.EmptyTransfer()
		if balance.Value.Cmp(decimal.Zero) > 0 {
			transfer.To = account.Account{
				Network: operation.Network,
				Address: balance.Address,
				Type:    types.NewAccountType(balance.Address),
			}
		} else {
			transfer.From = account.Account{
				Network: operation.Network,
				Address: balance.Address,
				Type:    types.NewAccountType(balance.Address),
			}
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
		transfer := operation.EmptyTransfer()

		acc, err := parser.accounts.Get(network, balance.Address)
		if err != nil {
			if !errors.Is(err, pg.ErrNoRows) {
				return nil, err
			}
			transfer.To = account.Account{
				Address: balance.Address,
				Type:    types.NewAccountType(balance.Address),
			}
			transfer.Amount = balance.Value.Abs()
		} else {
			tb, err := parser.repo.Get(network, contract, acc.ID, balance.TokenID)
			if err != nil {
				return nil, err
			}

			delta := balance.Value.Sub(tb.Balance)
			if delta.Cmp(decimal.Zero) > 0 {
				transfer.To = account.Account{
					Address: balance.Address,
					Type:    types.NewAccountType(balance.Address),
				}
			} else {
				transfer.From = account.Account{
					Address: balance.Address,
					Type:    types.NewAccountType(balance.Address),
				}
			}
			transfer.Amount = delta.Abs()
		}

		transfer.TokenID = balance.TokenID

		transfers = append(transfers, transfer)
	}

	return transfers, nil
}
