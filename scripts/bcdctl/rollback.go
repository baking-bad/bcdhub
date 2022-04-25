package main

import (
	"context"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/rollback"
)

type rollbackCommand struct {
	Level   int64  `short:"l" long:"level" description:"Level to rollback"`
	Network string `short:"n" long:"network" description:"Network"`
}

var rollbackCmd rollbackCommand

// Execute
func (x *rollbackCommand) Execute(_ []string) error {
	network := types.NewNetwork(x.Network)
	ctx, err := ctxs.Get(network)
	if err != nil {
		panic(err)
	}

	state, err := ctx.Blocks.Last()
	if err != nil {
		panic(err)
	}

	logger.Warning().Msgf("Do you want to rollback '%s' from %d to %d? (yes - continue. no - cancel)", network.String(), state.Level, x.Level)
	if !yes() {
		logger.Info().Msg("Cancelled")
		return nil
	}

	manager := rollback.NewManager(ctx.RPC, ctx.Searcher, ctx.Storage, ctx.Blocks, ctx.BigMapDiffs, ctx.Transfers)
	if err = manager.Rollback(context.Background(), ctx.StorageDB.DB, network, state, x.Level); err != nil {
		return err
	}
	logger.Info().Msg("Done")

	return nil
}
