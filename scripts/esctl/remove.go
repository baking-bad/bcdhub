package main

import (
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/rollback"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/pkg/errors"
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
	api := ctx.ES.GetAPI()
	options := []func(*esapi.IndicesDeleteRequest){
		api.Indices.Delete.WithAllowNoIndices(true),
	}

	resp, err := api.Indices.Delete(mappingNames, options...)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.IsError() {
		return errors.Errorf(resp.Status())
	}

	return nil
}
