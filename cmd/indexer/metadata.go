package main

import (
	"encoding/json"
	"fmt"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
)

func getMetadata(rpc *noderpc.NodeRPC, c *models.Contract, tag string) (map[string]string, error) {
	res := make(map[string]string)
	a, err := createMetadata(rpc, 0, c, tag)
	if err != nil {
		return nil, err
	}

	if c.Network == "mainnet" {
		res["babylon"] = a

		if c.Level < levelBabylon {
			a, err = createMetadata(rpc, levelBabylon-1, c, tag)
			if err != nil {
				return nil, err
			}
			res["alpha"] = a
		}
	} else {
		res["alpha"] = a
	}
	return res, nil
}

func createMetadata(rpc *noderpc.NodeRPC, level int64, c *models.Contract, tag string) (string, error) {
	contract, err := rpc.GetContractJSON(c.Address, level)
	if err != nil {
		return "", err
	}

	if contract.Get("spendable").Bool() {
		c.Tags = append(c.Tags, contractparser.SpendableTag)
	}

	args := contract.Get(fmt.Sprintf("script.code.#(prim==\"%s\").args", tag))
	if args.Exists() {
		a, err := contractparser.ParseMetadata(args)
		if err != nil {
			return "", nil
		}
		b, err := json.Marshal(a)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return "", fmt.Errorf("[createMetadata] Unknown tag '%s'", tag)
}

func saveMetadata(es *elastic.Elastic, rpc *noderpc.NodeRPC, c *models.Contract) error {
	storage, err := getMetadata(rpc, c, "storage")
	if err != nil {
		return err
	}
	parameter, err := getMetadata(rpc, c, "parameter")
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"parameter": parameter,
		"storage":   storage,
	}
	_, err = es.AddDocumentWithID(data, elastic.DocMetadata, c.Address)
	return err
}
