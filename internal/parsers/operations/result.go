package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

func parseOperationResult(data *noderpc.Operation) *operation.Result {
	result := data.GetResult()
	if result == nil {
		return &operation.Result{}
	}

	operationResult := operation.Result{
		Status:      result.Status,
		ConsumedGas: result.ConsumedGas,
	}
	if result.StorageSize != nil {
		operationResult.StorageSize = *result.StorageSize
	}
	if result.PaidStorageSizeDiff != nil {
		operationResult.PaidStorageSizeDiff = *result.PaidStorageSizeDiff
	}
	if len(result.Originated) > 0 {
		operationResult.Originated = result.Originated[0]
	}

	operationResult.AllocatedDestinationContract = data.Kind == consts.Origination
	if !operationResult.AllocatedDestinationContract && result.AllocatedDestinationContract != nil {
		operationResult.AllocatedDestinationContract = *result.AllocatedDestinationContract
	}
	errs, err := tezerrors.ParseArray(result.Errors)
	if err == nil {
		operationResult.Errors = errs
	}
	return &operationResult
}
