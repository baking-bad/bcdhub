package parsers

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
)

func createNewContract(es *elastic.Elastic, operation models.Operation, filesDirectory, protoSymLink string) (*models.Contract, error) {
	if operation.Kind != consts.Origination && operation.Kind != consts.Migration {
		return nil, fmt.Errorf("Invalid operation kind in computeContractMetrics: %s", operation.Kind)
	}
	contract := &models.Contract{
		ID:        strings.ReplaceAll(uuid.New().String(), "-", ""),
		Network:   operation.Network,
		Level:     operation.Level,
		Timestamp: operation.Timestamp,
		Manager:   operation.Source,
		Address:   operation.Destination,
		Balance:   operation.Balance,
		Delegate:  operation.Delegate,
	}

	err := computeMetrics(es, operation, filesDirectory, protoSymLink, contract)
	return contract, err
}

func computeMetrics(es *elastic.Elastic, operation models.Operation, filesDirectory, protoSymLink string, contract *models.Contract) error {
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
	if err := saveToFile(operation.Script, contract, filesDirectory, protoSymLink); err != nil {
		return err
	}

	return saveMetadata(es, operation.Script, filesDirectory, protoSymLink, contract)
}

func saveToFile(script gjson.Result, c *models.Contract, filesDirectory, protoSymLink string) error {
	filePath := fmt.Sprintf("%s/contracts/%s/%s_%s.json", filesDirectory, c.Network, c.Address, protoSymLink)
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
