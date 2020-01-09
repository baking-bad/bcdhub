package main

import (
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
)

func computeMetrics(rpc *noderpc.NodeRPC, c *models.Contract) error {
	contract, err := rpc.GetContract(c.Address)
	if err != nil {
		return err
	}

	script, err := contractparser.New(contract)
	if err != nil {
		return err
	}
	script.Parse()

	c.Language = script.Language()
	c.Hash = []string{
		script.Code.Parameter.Hash,
		script.Code.Hash,
		script.Code.Storage.Hash,
	}
	c.FailStrings = script.Code.FailStrings.Values()
	c.Primitives = script.Code.Primitives.Values()
	c.Annotations = script.Code.Annotations.Values()
	c.Entrypoints = script.Code.Parameter.Entrypoints()
	c.Tags = script.Tags.Values()

	c.Hardcoded = script.HardcodedAddresses

	return nil
}
