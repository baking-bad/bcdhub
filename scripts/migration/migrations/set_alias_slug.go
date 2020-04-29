package migrations

import "github.com/baking-bad/bcdhub/internal/logger"

// SetAliasSlug - migration that set `slug` at Alias in db
type SetAliasSlug struct{}

// Description -
func (m *SetAliasSlug) Description() string {
	return "set `slug` at Alias in db"
}

// Do - migrate function
func (m *SetAliasSlug) Do(ctx *Context) error {
	for _, network := range ctx.Config.Migrations.Networks {
		all, err := ctx.DB.GetAliases(network)
		if err != nil {
			return err
		}

		for i := range all {
			if err := ctx.DB.CreateOrUpdateAlias(&all[i]); err != nil {
				return err
			}
		}
	}

	logger.Info("Done")
	return nil
}
