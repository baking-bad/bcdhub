package indexer

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

func (bi *BoostIndexer) findNewContracts(operations []models.Operation) ([]models.Contract, error) {
	contracts := make([]models.Contract, 0)

	for _, operation := range operations {
		if operation.Kind != consts.Origination {
			continue
		}

		contract, err := bi.createNewContract(operation)
		if err != nil {
			return nil, err
		}
		contracts = append(contracts, contract)
	}

	return contracts, nil
}

func (bi *BoostIndexer) createNewContract(operation models.Operation) (models.Contract, error) {
	contract := models.Contract{
		ID:        strings.ReplaceAll(uuid.New().String(), "-", ""),
		Network:   operation.Network,
		Level:     operation.Level,
		Timestamp: operation.Timestamp,
		Manager:   operation.Source,
		Address:   operation.Destination,
		Balance:   operation.Balance,
		Delegate:  operation.Delegate,
	}
	err := bi.computeMetrics(operation, bi.filesDirectory, &contract)
	return contract, err
}

func (bi *BoostIndexer) computeMetrics(operation models.Operation, filesDirectory string, contract *models.Contract) error {
	if operation.Kind != consts.Origination {
		return fmt.Errorf("Invalid operation kind in computeContractMetrics: %s", operation.Kind)
	}

	script, err := contractparser.New(operation.Script)
	if err != nil {
		return fmt.Errorf("contractparser.New: %v", err)
	}
	script.Parse()

	contract.Language = script.Language()
	contract.FailStrings = script.Code.FailStrings.Values()
	contract.Primitives = script.Code.Primitives.Values()
	contract.Annotations = script.Code.Annotations.Values()
	contract.Tags = script.Tags.Values()

	contract.Hardcoded = script.HardcodedAddresses.Values()

	if err := computeFingerprint(operation.Script, contract); err != nil {
		return err
	}

	if err := saveToFile(operation.Script, contract, 0, filesDirectory); err != nil {
		return err
	}

	// Save contract code before babylon in mainnet
	if contract.Level < consts.LevelBabylon && contract.Network == consts.Mainnet {
		alphaContract, err := contractparser.GetContract(bi.rpc, contract.Address, contract.Network, contract.Level, filesDirectory)
		if err != nil {
			return err
		}
		if err := saveToFile(alphaContract, contract, contract.Level, filesDirectory); err != nil {
			return err
		}
	}

	return saveMetadata(bi.es, bi.rpc, contract, filesDirectory)
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
