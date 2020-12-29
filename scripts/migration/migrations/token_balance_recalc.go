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

	if !helpers.IsContract(address) {
		return errors.Errorf("Invalid contract address: `%s`", address)
	}

	logger.Info("Removing token balance entities....")
	if err := ctx.ES.DeleteByContract([]string{elastic.DocTokenBalances}, network, address); err != nil {
		return err
	}

	logger.Info("Receiving transfers....")
	updates := make([]*models.Transfer, 0)

	var lastID string
	for {
		transfers, err := ctx.ES.GetTransfers(elastic.GetTransfersContext{
			Network:   network,
			Contracts: []string{address},
			LastID:    lastID,
			TokenID:   -1,
		})
		if err != nil {
			return err
		}
		if len(transfers.Transfers) == 0 {
			break
		}
		for i := range transfers.Transfers {
			updates = append(updates, &transfers.Transfers[i])
		}
		lastID = transfers.LastID
	}

	logger.Info("Saving...")
	return elastic.CreateTokenBalanceUpdates(ctx.ES, updates)
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
