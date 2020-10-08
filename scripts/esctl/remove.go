package main

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/rollback"
)

type removeCommand struct {
	Network string `short:"n" long:"network" description:"Network"`
}

var removeCmd removeCommand

// Execute
func (x *removeCommand) Execute(args []string) error {
	state, err := ctx.ES.GetLastBlock(x.Network)
	if err != nil {
		panic(err)
	}

	logger.Warning("Do you want to remove data of '%s' network? (yes - continue. no - cancel)", state.Network)
	if !yes() {
		logger.Success("Cancelled")
		return nil
	}

	if err = rollback.Remove(ctx.ES, x.Network, ctx.Config.Share.Path); err != nil {
		return err
	}

	logger.Success("Done")
	return nil
}

type deleteIndicesCommand struct {
	Network string `short:"n" long:"network" description:"Network"`
}

var deleteIndicesCmd deleteIndicesCommand

// Execute
func (x *deleteIndicesCommand) Execute(args []string) error {
	return ctx.ES.DeleteIndices(mappingNames)
}
