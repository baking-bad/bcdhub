package migrations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic/core"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/schollz/progressbar/v3"
)

// SetAliases - migration that set aliases for operations (source, destination, destination) and contracts (address, delegate)
type SetAliases struct{}

// Key -
func (m *SetAliases) Key() string {
	return "set_aliases"
}

// Description -
func (m *SetAliases) Description() string {
	return "set aliases for operations (source, destination, delegate) and contracts (address, delegate)"
}

// Do - migrate function
func (m *SetAliases) Do(ctx *config.Context) error {
	logger.Info("Receiving aliases for %s...", consts.Mainnet)

	aliases, err := ctx.TZIP.GetAliasesMap(consts.Mainnet)
	if err != nil {
		return err
	}
	logger.Info("Received %d aliases", len(aliases))

	if len(aliases) == 0 {
		return nil
	}

	bar := progressbar.NewOptions(len(aliases), progressbar.OptionSetPredictTime(false), progressbar.OptionClearOnFinish(), progressbar.OptionShowCount())
	for address, alias := range aliases {
		if err := bar.Add(1); err != nil {
			return err
		}

		query := core.NewQuery().Query(
			core.Bool(
				core.Filter(
					core.Term("network", consts.Mainnet),
					core.Bool(
						core.Should(
							core.MatchPhrase("address", address),
							core.MatchPhrase("source", address),
							core.MatchPhrase("destination", address),
							core.MatchPhrase("delegate", address),
						),
						core.MinimumShouldMatch(1),
					),
				),
			),
		).Add(
			core.Item{
				"script": core.Item{
					"source": `
					if (ctx._index == "contract") {
						if (ctx._source.address == params.address) {
							ctx._source.alias = params.alias
						}

						if (ctx._source.delegate == params.address) {
							ctx._source.delegate_alias = params.alias
						}
					} else if (ctx._index == 'operation') {
						if (ctx._source.source == params.address) {
							ctx._source.source_alias = params.alias
						}

						if (ctx._source.destination == params.address) {
							ctx._source.destination_alias = params.alias
						}

						if (ctx._source.delegate == params.address) {
							ctx._source.delegate_alias = params.alias
						}
					}`,
					"lang": "painless",
					"params": core.Item{
						"alias":   alias,
						"address": address,
					},
				},
			},
		)

		if err := ctx.Storage.UpdateByQueryScript(
			[]string{models.DocOperations, models.DocContracts},
			query,
		); err != nil {
			return fmt.Errorf("%s %s %w", address, alias, err)
		}
	}

	return nil
}
