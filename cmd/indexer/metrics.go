package main

import (
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
)

func computeMetrics(rpc *noderpc.NodeRPC, c *models.Contract) error {
	bScript, err := rpc.GetContractScriptBytes(c.Address)
	if err != nil {
		return err
	}

	script, err := contractparser.New(bScript)
	if err != nil {
		return err
	}
	if err := script.Parse(); err != nil {
		return err
	}

	c.Language = script.Language()
	c.Kind = script.Kind()
	c.HashCode = script.Code.Hash

	c.Tags = make([]string, 0)
	for tag := range script.Tags {
		c.Tags = append(c.Tags, tag)
	}

	c.Hardcoded = script.HardcodedAddresses

	return nil
}
