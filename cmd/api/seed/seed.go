package seed

import (
	"github.com/baking-bad/bcdhub/cmd/api/handlers"
	"github.com/baking-bad/bcdhub/internal/config"
	"github.com/baking-bad/bcdhub/internal/database"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/models/tzip"
)

// Run -
func Run(ctx *handlers.Context, seed config.SeedConfig) error {
	// 1. seed user
	user := database.User{
		Login:     seed.User.Login,
		Name:      seed.User.Name,
		AvatarURL: seed.User.AvatarURL,
	}

	if err := ctx.DB.GetOrCreateUser(&user, ""); err != nil {
		return err
	}

	ctx.OAUTH.UserID = user.ID

	// 2. seed subscriptions
	for _, sub := range seed.Subscriptions {
		subscription := database.Subscription{
			UserID:    user.ID,
			Address:   sub.Address,
			Network:   types.NewNetwork(sub.Network),
			Alias:     sub.Alias,
			WatchMask: sub.WatchMask,
		}

		if err := ctx.DB.UpsertSubscription(&subscription); err != nil {
			return err
		}
	}

	// 3. seed aliases
	aliasModels := make([]models.Model, 0)
	for _, a := range seed.Aliases {
		aliasModels = append(aliasModels, &tzip.TZIP{
			TZIP16: tzip.TZIP16{
				Name: a.Alias,
			},
			Network: types.NewNetwork(a.Network),
			Address: a.Address,
		})
	}
	if err := ctx.Storage.Save(aliasModels); err != nil {
		return err
	}

	// 4. seed accounts
	for _, a := range seed.Accounts {
		account := database.Account{
			UserID:        user.ID,
			PrivateKey:    a.PrivateKey,
			PublicKeyHash: a.PublicKeyHash,
			Network:       types.NewNetwork(a.Network),
		}

		if err := ctx.DB.GetOrCreateAccount(&account); err != nil {
			return err
		}
	}
	return nil
}
