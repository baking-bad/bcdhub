package migrations

import (
	"log"
	"time"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
	"github.com/baking-bad/bcdhub/internal/tzkt"
	"github.com/schollz/progressbar/v3"
)

// Aliases -
type Aliases struct{}

// Key -
func (m *Aliases) Key() string {
	return "aliases"
}

// Description -
func (m *Aliases) Description() string {
	return "fill aliases from TzKT"
}

// Do - migrate function
func (m *Aliases) Do(ctx *config.Context) error {
	logger.Info("Starting fill aliases...")

	cfg := ctx.Config.TzKT[consts.Mainnet]
	timeout := time.Duration(cfg.Timeout) * time.Second

	api := tzkt.NewTzKT(cfg.URI, timeout)
	logger.Info("TzKT API initialized")

	aliases, err := api.GetAliases()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("Got %d aliases from tzkt api", len(aliases))
	logger.Info("Saving aliases to elastic...")

	newModels := make([]elastic.Model, 0)
	bar := progressbar.NewOptions(len(aliases), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for address, alias := range aliases {
		if err := bar.Add(1); err != nil {
			return err
		}

		item := models.TZIP{
			Network: consts.Mainnet,
			Address: address,
			Slug:    helpers.Slug(alias),
			TZIP16: tzip.TZIP16{
				Name: alias,
			},
		}
		if err := ctx.ES.GetByID(&item); err == nil {
			continue
		} else if !elastic.IsRecordNotFound(err) {
			log.Println(err)
			return err
		}
		newModels = append(newModels, &item)
	}
	return ctx.ES.BulkInsert(newModels)
}
