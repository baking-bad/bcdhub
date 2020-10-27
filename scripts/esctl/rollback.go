package main

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/rollback"
)

type rollbackCommand struct {
	Level   int64  `short:"l" long:"level" description:"Level to rollback"`
	Network string `short:"n" long:"network" description:"Network"`
}

var rollbackCmd rollbackCommand

// Execute
func (x *rollbackCommand) Execute(_ []string) error {
	state, err := ctx.ES.GetLastBlock(x.Network)
	if err != nil {
		panic(err)
	}

	logger.Warning("Do you want to rollback '%s' from %d to %d? (yes - continue. no - cancel)", state.Network, state.Level, x.Level)
	if !yes() {
		logger.Success("Cancelled")
		return nil
	}

	rpc, err := ctx.GetRPC(state.Network)
	if err != nil {
		panic(err)
	}

	manager := rollback.NewManager(ctx.ES, ctx.MQ, rpc, ctx.SharePath)
	if err = manager.Rollback(state, x.Level); err != nil {
		return err
	}
	logger.Success("Done")

	return nil
}
