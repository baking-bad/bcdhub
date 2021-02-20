package migrations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/logger"
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

		if err := ctx.Storage.SetAlias(consts.Mainnet, address, alias); err != nil {
			return err
		}
	}

	return nil
}
