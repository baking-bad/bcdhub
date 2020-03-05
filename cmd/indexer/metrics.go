package main

import (
	"fmt"
	"os"
	"path"

	"github.com/aopoltorzhicky/bcdhub/internal/contractparser"
	"github.com/aopoltorzhicky/bcdhub/internal/contractparser/consts"
	"github.com/aopoltorzhicky/bcdhub/internal/elastic"
	"github.com/aopoltorzhicky/bcdhub/internal/models"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/tidwall/gjson"
)

func computeMetrics(rpc noderpc.Pool, es *elastic.Elastic, c *models.Contract, filesDirectory string) error {
	contract, err := contractparser.GetContract(rpc, c.Address, c.Network, 0, filesDirectory)
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

	if err := saveToFile(contract, c, 0, filesDirectory); err != nil {
		return err
	}

	// Save contract code before babylon in mainnet
	if c.Level < consts.LevelBabylon && c.Network == consts.Mainnet {
		alphaContract, err := contractparser.GetContract(rpc, c.Address, c.Network, c.Level, filesDirectory)
		if err != nil {
			return err
		}
		if err := saveToFile(alphaContract, c, c.Level, filesDirectory); err != nil {
			return err
		}
	}

	return saveMetadata(es, rpc, c, filesDirectory)
}

func saveToFile(script gjson.Result, c *models.Contract, level int64, filesDirectory string) error {
	var postfix string
	if c.Network == consts.Mainnet {
		if level < consts.LevelBabylon && level != 0 {
			postfix = "_alpha"
		} else {
			postfix = "_babylon"
		}
	}

	filePath := fmt.Sprintf("%s/contracts/%s/%s%s.json", filesDirectory, c.Network, c.Address, postfix)
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
