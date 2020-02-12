package main

import (
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
)

func computeMetrics(rpc *noderpc.NodeRPC, es *elastic.Elastic, c *models.Contract) error {
	contract, err := rpc.GetScriptJSON(c.Address, 0)
	if err != nil {
		return err
	}

	script, err := contractparser.New(contract)
	if err != nil {
		return fmt.Errorf("contractparser.New: %v", err)
	}
	script.Parse()

	c.Language = script.Language()
	c.FailStrings = script.Code.FailStrings.Values()
	c.Primitives = script.Code.Primitives.Values()
	c.Annotations = script.Code.Annotations.Values()
	c.Tags = script.Tags.Values()

	c.Hardcoded = script.HardcodedAddresses.Values()

	return saveMetadata(es, rpc, c)
}
