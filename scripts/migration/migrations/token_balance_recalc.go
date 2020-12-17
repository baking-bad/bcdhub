package migrations

import (
	"strings"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
)

// TokenBalanceRecalc - migration that recalculates token balances for contract
type TokenBalanceRecalc struct{}

// Key -
func (m *TokenBalanceRecalc) Key() string {
	return "token_balance_recalc"
}

// Description -
func (m *TokenBalanceRecalc) Description() string {
	return "recalculates token balances for contract"
}

// Do - migrate function
func (m *TokenBalanceRecalc) Do(ctx *config.Context) error {
	network, err := ask("Enter network (if empty - mainnet):")
	if err != nil {
		return err
	}
	if network == "" {
		network = "mainnet"
	}

	if !helpers.StringInArray(network, ctx.Config.Scripts.Networks) {
		return errors.Errorf("Invalid network: `%s`. Availiable values: %s", network, strings.Join(ctx.Config.Scripts.Networks, ","))
	}

	address, err := ask("Enter contract address (required):")
	if err != nil {
		return err
	}
	if !helpers.IsContract(address) {
		return errors.Errorf("Invalid contract address: `%s`", address)
	}

	logger.Info("Removing token balance entities....")
	if err := ctx.ES.DeleteByContract([]string{elastic.DocTokenBalances}, network, address); err != nil {
		return err
	}

	logger.Info("Receiving new balances....")
	balances, err := ctx.ES.GetBalances(network, address, 0)
	if err != nil {
		return err
	}

	updates := make([]*models.TokenBalance, 0)
	for key, balance := range balances {
		updates = append(updates, &models.TokenBalance{
			Network:  network,
			Address:  key.Address,
			Contract: address,
			TokenID:  key.TokenID,
			Balance:  balance,
		})
	}

	logger.Info("Saving...")
	return ctx.ES.UpdateTokenBalances(updates)
}
