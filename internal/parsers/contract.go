package parsers

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/baking-bad/bcdhub/internal/contractparser"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/contractparser/kinds"
	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/elastic"
	"github.com/baking-bad/bcdhub/internal/helpers"
	"github.com/baking-bad/bcdhub/internal/metrics"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

func createNewContract(es elastic.IElastic, interfaces map[string]kinds.ContractKind, operation models.Operation, filesDirectory, protoSymLink string) ([]elastic.Model, error) {
	if !helpers.StringInArray(operation.Kind, []string{
		consts.Origination, consts.OriginationNew, consts.Migration,
	}) {
		return nil, errors.Errorf("Invalid operation kind in computeContractMetrics: %s", operation.Kind)
	}
	contract := models.Contract{
		ID:         helpers.GenerateID(),
		Network:    operation.Network,
		Level:      operation.Level,
		Timestamp:  operation.Timestamp,
		Manager:    operation.Source,
		Address:    operation.Destination,
		Balance:    operation.Amount,
		Delegate:   operation.Delegate,
		LastAction: models.BCDTime{Time: operation.Timestamp},
		TxCount:    1,
	}

	if err := computeMetrics(operation, interfaces, filesDirectory, protoSymLink, &contract); err != nil {
		return nil, err
	}

	metadata, err := createMetadata(operation.Script, protoSymLink, &contract)
	if err != nil {
		return nil, err
	}

	if _, err := es.AddDocumentWithID(metadata, elastic.DocMetadata, contract.Address); err != nil {
		return nil, err
	}

	upgradable, err := isUpgradable(*metadata, protoSymLink)
	if err != nil {
		return nil, err
	}

	if upgradable {
		contract.Tags = append(contract.Tags, consts.UpgradableTag)
	}

	if err := setEntrypoints(*metadata, protoSymLink, &contract); err != nil {
		return nil, err
	}

	return []elastic.Model{&contract}, nil
}

func computeMetrics(operation models.Operation, interfaces map[string]kinds.ContractKind, filesDirectory, protoSymLink string, contract *models.Contract) error {
	script, err := contractparser.New(operation.Script)
	if err != nil {
		return errors.Errorf("contractparser.New: %v", err)
	}
	script.Parse(interfaces)

	lang, err := script.Language()
	if err != nil {
		return errors.Errorf("script.Language: %v", err)
	}

	contract.Language = lang
	contract.Hash = script.Code.Hash
	contract.FailStrings = script.Code.FailStrings.Values()
	contract.Annotations = script.Annotations.Values()
	contract.Tags = script.Tags.Values()
	contract.Hardcoded = script.HardcodedAddresses.Values()

	if err := metrics.SetFingerprint(operation.Script, contract); err != nil {
		return err
	}
	return saveToFile(operation.Script, contract, filesDirectory, protoSymLink)
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
	} else if err != nil {
		return err
	}
	return nil
}

func isUpgradable(metadata models.Metadata, protoSymLink string) (bool, error) {
	parameter := metadata.Parameter[protoSymLink]
	storage := metadata.Storage[protoSymLink]

	var paramMeta meta.Metadata
	if err := json.Unmarshal([]byte(parameter), &paramMeta); err != nil {
		return false, errors.Errorf("Invalid parameter metadata: %v", err)
	}

	var storageMeta meta.Metadata
	if err := json.Unmarshal([]byte(storage), &storageMeta); err != nil {
		return false, errors.Errorf("Invalid parameter metadata: %v", err)
	}

	for _, p := range paramMeta {
		if p.Type != consts.LAMBDA {
			continue
		}

		for _, s := range storageMeta {
			if s.Type != consts.LAMBDA {
				continue
			}

			if p.Parameter == s.Parameter {
				return true, nil
			}
		}
	}

	return false, nil
}

func setEntrypoints(metadata models.Metadata, protoSymLink string, contract *models.Contract) error {
	var parameterMetadata meta.Metadata
	if err := json.Unmarshal([]byte(metadata.Parameter[protoSymLink]), &parameterMetadata); err != nil {
		return err
	}
	entrypoints, err := parameterMetadata.GetEntrypoints()
	if err != nil {
		return err
	}
	contract.Entrypoints = make([]string, len(entrypoints))
	for i := range entrypoints {
		contract.Entrypoints[i] = entrypoints[i].Name
	}
	return nil
}
