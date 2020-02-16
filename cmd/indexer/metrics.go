package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

func computeMetrics(rpc *noderpc.NodeRPC, es *elastic.Elastic, c *models.Contract) error {
	contract, err := getContractCode(rpc, c.Address, c.Network)
	if err != nil {
		return err
	}

	s := contract.Get("script")
	script, err := contractparser.New(s)
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

	if contract.Get("spendable").Bool() {
		c.Tags = append(c.Tags, consts.SpendableTag)
	}

	if err := computeFingerprint(s, c); err != nil {
		return err
	}

	if err := saveToFile(contract, c); err != nil {
		return err
	}

	return saveMetadata(es, rpc, c, s)
}

func saveToFile(script gjson.Result, c *models.Contract) error {
	filePath := fmt.Sprintf("%s/contracts/%s/%s.json", filesDirectory, c.Network, c.Address)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		d := path.Dir(filePath)
		if _, err := os.Stat(d); os.IsNotExist(err) {
			if err := os.MkdirAll(d, os.ModePerm); err != nil {
				return err
			}
		}

		f, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := f.WriteString(script.String()); err != nil {
			return err
		}
	}
	return nil
}

func getContractCode(rpc *noderpc.NodeRPC, address, network string) (gjson.Result, error) {
	filePath := fmt.Sprintf("%s/contracts/%s/%s.json", filesDirectory, network, address)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return rpc.GetContractJSON(address, 0)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return gjson.Result{}, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.ParseBytes(data), nil
}
