package migrations

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/tzkt"
	"github.com/go-pg/pg/v10"
	"github.com/schollz/progressbar/v3"
)

// GetAliases -
type GetAliases struct{}

// Key -
func (m *GetAliases) Key() string {
	return "get_aliases"
}

// Description -
func (m *GetAliases) Description() string {
	return "get aliases from TzKT"
}

// Do - migrate function
func (m *GetAliases) Do(ctx *config.Context) error {
	logger.Info().Msg("Starting get aliases...")

	cfg := ctx.Config.TzKT["mainnet"]
	timeout := time.Duration(cfg.Timeout) * time.Second

	api := tzkt.NewTzKT(cfg.URI, timeout)
	logger.Info().Msg("TzKT API initialized")

	aliases, err := api.GetAliases()
	if err != nil {
		return err
	}
	logger.Info().Msgf("Got %d aliases from tzkt api", len(aliases))
	logger.Info().Msg("Saving aliases...")

	return ctx.StorageDB.DB.RunInTransaction(context.Background(), func(tx *pg.Tx) error {
		bar := progressbar.NewOptions(len(aliases), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
		for address, alias := range aliases {
			if err := bar.Add(1); err != nil {
				return err
			}

			acc := account.Account{
				Network: types.Mainnet,
				Address: address,
				Type:    types.NewAccountType(address),
				Alias:   alias,
			}

			if err := acc.Save(tx); err != nil {
				return err
			}
		}
		return nil
	})
}
