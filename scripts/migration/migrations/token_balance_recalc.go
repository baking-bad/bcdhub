package migrations

import (
	"context"
	"strings"

	"github.com/baking-bad/bcdhub/internal/bcd"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tokenbalance"
	"github.com/baking-bad/bcdhub/internal/models/types"
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
	recalcAllEvents, err := ask("Recalc all contract with events (if empty - yes):")
	if err != nil {
		return err
	}
	if recalcAllEvents == "" {
		return m.RecalcAllContractEvents(ctx)
	}

	network, err := ask("Enter network (if empty - mainnet):")
	if err != nil {
		return err
	}
	if network == "" {
		network = "mainnet"
	}

	address, err := ask("Enter contract address (required):")
	if err != nil {
		return err
	}

	return m.Recalc(ctx, network, address)
}

// Recalc -
func (m *TokenBalanceRecalc) Recalc(ctx *config.Context, network, address string) error {
	if !helpers.StringInArray(network, ctx.Config.Scripts.Networks) {
		return errors.Errorf("Invalid network: `%s`. Availiable values: %s", network, strings.Join(ctx.Config.Scripts.Networks, ","))
	}

	if !bcd.IsContract(address) {
		logger.Error().Msgf("Invalid contract address: `%s`", address)
		return nil
	}

	typ := types.NewNetwork(network)

	logger.Info().Msg("Removing token balance entities....")
	if err := ctx.Storage.DeleteByContract(typ, []string{models.DocTokenBalances}, address); err != nil {
		return err
	}

	balances, err := ctx.Transfers.CalcBalances(typ, address)
	if err != nil {
		return err
	}
	logger.Info().Msgf("Received %d balances", len(balances))

	updates := make([]models.Model, 0)
	for _, balance := range balances {
		updates = append(updates, &tokenbalance.TokenBalance{
			Network:  typ,
			Address:  balance.Address,
			Contract: address,
			TokenID:  balance.TokenID,
			Balance:  balance.Balance,
			IsLedger: true,
		})
	}

	logger.Info().Msg("Saving...")
	return ctx.Storage.Save(context.Background(), updates)
}

// DoBatch -
func (m *TokenBalanceRecalc) DoBatch(ctx *config.Context, contracts map[string]string) error {
	for address, network := range contracts {
		if err := m.Recalc(ctx, network, address); err != nil {
			return err
		}
	}

	return nil
}

// RecalcAllContractEvents -
func (m *TokenBalanceRecalc) RecalcAllContractEvents(ctx *config.Context) error {
	tzips, err := ctx.TZIP.GetWithEvents(0)
	if err != nil {
		return err
	}

	for _, tzip := range tzips {
		logger.Info().Msgf("Starting %s %s", tzip.Network, tzip.Address)
		if err := m.Recalc(ctx, tzip.Network.String(), tzip.Address); err != nil {
			return err
		}
	}

	return nil
}
