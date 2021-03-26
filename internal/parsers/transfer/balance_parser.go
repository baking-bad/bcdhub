package transfer

import (
	"math"

	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/transfer"
	tbParser "github.com/baking-bad/bcdhub/internal/parsers/tokenbalance"
)

// DefaultBalanceParser -
type DefaultBalanceParser struct {
	repo models.GeneralRepository
}

// NewDefaultBalanceParser -
func NewDefaultBalanceParser(repo models.GeneralRepository) *DefaultBalanceParser {
	return &DefaultBalanceParser{repo}
}

// Parse -
func (parser *DefaultBalanceParser) Parse(balances []tbParser.TokenBalance, operation operation.Operation) ([]*transfer.Transfer, error) {
	transfers := make([]*transfer.Transfer, 0)
	for _, balance := range balances {
		transfer := transfer.EmptyTransfer(operation)
		if balance.Value > 0 {
			transfer.To = balance.Address
		} else {
			transfer.From = balance.Address
		}
		transfer.Amount = math.Abs(balance.Value)
		transfer.TokenID = balance.TokenID

		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

// ParseBalances -
func (parser *DefaultBalanceParser) ParseBalances(network, contract string, balances []tbParser.TokenBalance, operation operation.Operation) ([]*transfer.Transfer, error) {
	transfers := make([]*transfer.Transfer, 0)
	for _, balance := range balances {
		transfer := transfer.EmptyTransfer(operation)

		tb := tokenbalance.TokenBalance{
			Network:  network,
			Contract: contract,
			Address:  balance.Address,
			TokenID:  balance.TokenID,
		}
		if err := parser.repo.GetByID(&tb); err != nil {
			if !parser.repo.IsRecordNotFound(err) {
				return nil, err
			}

			tb.Balance = 0
		}

		delta := balance.Value - tb.Balance

		if delta > 0 {
			transfer.To = balance.Address
		} else {
			transfer.From = balance.Address
		}

		transfer.Amount = math.Abs(delta)
		transfer.TokenID = balance.TokenID

		transfers = append(transfers, transfer)
	}

	return transfers, nil
}
