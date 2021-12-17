package migrations

import (
	"context"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/tzkt"
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

	newModels := make([]models.Model, 0)
	updated := make([]models.Model, 0)

	bar := progressbar.NewOptions(len(aliases), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for address, alias := range aliases {
		if err := bar.Add(1); err != nil {
			return err
		}

		item, err := ctx.TZIP.Get(types.Mainnet, address)
		switch {
		case err == nil:
			item.Name = alias
			item.Slug = helpers.Slug(alias)

			updated = append(updated, item)
		case ctx.Storage.IsRecordNotFound(err):
			newModels = append(newModels, &tzip.TZIP{
				Network: types.Mainnet,
				Address: address,
				Slug:    helpers.Slug(alias),
				TZIP16: tzip.TZIP16{
					Name: alias,
				},
			})
		default:
			return err
		}
	}
	if err := ctx.Storage.Save(context.Background(), updated); err != nil {
		return err
	}

	return ctx.Storage.Save(context.Background(), newModels)
}
