package operations

import (
	"fmt"

	"github.com/baking-bad/bcdhub/internal/contractparser/cerrors"
	"github.com/baking-bad/bcdhub/internal/contractparser/consts"
	"github.com/baking-bad/bcdhub/internal/models"
	"github.com/tidwall/gjson"
)

// Result -
type Result struct {
	root string
}

// NewResult -
func NewResult(root string) Result {
	return Result{root}
}

// Parse -
func (r Result) Parse(data gjson.Result) models.OperationResult {
	if r.root != "" {
		r.root = fmt.Sprintf("%s.", r.root)
	}
	result := models.OperationResult{
		Status:                       data.Get(r.root + "status").String(),
		ConsumedGas:                  data.Get(r.root + "consumed_gas").Int(),
		StorageSize:                  data.Get(r.root + "storage_size").Int(),
		PaidStorageSizeDiff:          data.Get(r.root + "paid_storage_size_diff").Int(),
		Originated:                   data.Get(r.root + "originated_contracts.0").String(),
		AllocatedDestinationContract: data.Get(r.root+"allocated_destination_contract").Bool() || data.Get("kind").String() == consts.Origination,
	}
	err := data.Get(r.root + "errors")
	result.Errors = cerrors.ParseArray(err)
	return result
}
