package operations

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/baking-bad/bcdhub/internal/contractparser/meta"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// Metadata -
type Metadata struct {
	Result         models.OperationResult
	BalanceUpdates []models.BalanceUpdate
}

func parseMetadata(item gjson.Result) *Metadata {
	path := "metadata.operation_result"
	if !item.Get(path).Exists() {
		path = "result"
		if !item.Get(path).Exists() {
			return nil
		}
	}

	return &Metadata{
		BalanceUpdates: NewBalanceUpdate(path).Parse(item),
		Result:         NewResult(path).Parse(item),
	}
}

func readJSONFile(name string) (gjson.Result, error) {
	bytes, err := ioutil.ReadFile(name)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.ParseBytes(bytes), nil
}

func readTestMetadata(address string) (*meta.ContractMetadata, error) {
	bytes, err := ioutil.ReadFile(fmt.Sprintf("./data/metadata/%s.json", address))
	if err != nil {
		return nil, err
	}
	var metadata meta.ContractMetadata
	err = json.Unmarshal(bytes, &metadata)
	return &metadata, err
}
