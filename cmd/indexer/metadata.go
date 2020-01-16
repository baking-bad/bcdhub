package main

import (
	"encoding/json"
	"fmt"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
)

func getMetadatas(rpc *noderpc.NodeRPC, c models.Contract) (map[string]string, error) {
	res := make(map[string]string)
	a, err := createMetadata(rpc, 0, c.Address)
	if err != nil {
		return nil, err
	}

	if c.Network == "mainnet" {
		res["babylon"] = a

		if c.Level < levelBabylon {
			a, err = createMetadata(rpc, levelBabylon-1, c.Address)
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

func createMetadata(rpc *noderpc.NodeRPC, level int64, address string) (string, error) {
	data, err := rpc.GetScriptJSON(address, level)
	if err != nil {
		return "", err
	}
	for _, c := range data.Get("code").Array() {
		if c.Get("prim").String() == "storage" {
			a, err := contractparser.ParseMetadata(c.Get("args"))
			if err != nil {
				return "", nil
			}

			b, err := json.Marshal(a)
			if err != nil {
				return "", err
			}
			return string(b), nil
		}
	}
	return "", fmt.Errorf("[createMetadata] Invalid code structure")
}

func saveMetadatas(es *elastic.Elastic, rpc *noderpc.NodeRPC, c models.Contract) error {
	data, err := getMetadatas(rpc, c)
	if err != nil {
		return err
	}
	_, err = es.AddDocumentWithID(data, elastic.DocMetadata, c.Address)
	return err
}
