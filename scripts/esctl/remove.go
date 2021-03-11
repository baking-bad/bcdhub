package main

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/baking-bad/bcdhub/internal/rollback"
)

type removeCommand struct {
	Network string `short:"n" long:"network" description:"Network"`
}

var removeCmd removeCommand

// Execute
func (x *removeCommand) Execute(_ []string) error {
	state, err := ctx.Blocks.Last(x.Network)
	if err != nil {
		panic(err)
	}

	logger.Warning("Do you want to remove data of '%s' network? (yes - continue. no - cancel)", state.Network)
	if !yes() {
		logger.Info("Cancelled")
		return nil
	}

	if err = rollback.Remove(ctx.Storage, ctx.Contracts, x.Network); err != nil {
		return err
	}

	logger.Info("Done")
	return nil
}

type deleteIndicesCommand struct {
	Network string `short:"n" long:"network" description:"Network"`
}

var deleteIndicesCmd deleteIndicesCommand

// Execute
func (x *deleteIndicesCommand) Execute(_ []string) error {
	return ctx.Storage.DeleteIndices(models.AllDocuments())
}
