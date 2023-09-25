package operations

import (
	"github.com/baking-bad/bcdhub/internal/bcd/consts"
	"github.com/baking-bad/bcdhub/internal/bcd/tezerrors"
	"github.com/baking-bad/bcdhub/internal/models/account"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
	"github.com/baking-bad/bcdhub/internal/noderpc"
)

func parseOperationResult(data noderpc.Operation, tx *operation.Operation) {
	result := data.GetResult()
	if result == nil {
		return
	}

	tx.Status = types.NewOperationStatus(result.Status)

	if result.ConsumedMilligas != nil {
		tx.ConsumedGas = *result.ConsumedMilligas
	} else {
		tx.ConsumedGas = result.ConsumedGas * 100
	}

	if result.StorageSize != nil {
		tx.StorageSize = *result.StorageSize
	}
	if result.PaidStorageSizeDiff != nil {
		tx.PaidStorageSizeDiff = *result.PaidStorageSizeDiff
	}
	if len(result.Originated) > 0 {
		tx.Destination = account.Account{
			Address: result.Originated[0],
			Type:    types.AccountTypeContract,
			Level:   tx.Level,
		}
	}

	if len(result.OriginatedRollup) > 0 {
		tx.Destination = account.Account{
			Address: result.OriginatedRollup,
			Type:    types.AccountTypeRollup,
			Level:   tx.Level,
		}
	}

	tx.AllocatedDestinationContract = data.Kind == consts.Origination
	if result.AllocatedDestinationContract != nil {
		tx.AllocatedDestinationContract = *result.AllocatedDestinationContract
	}

	if errs, err := tezerrors.ParseArray(result.Errors); err == nil {
		tx.Errors = errs
	}

	if tx.IsApplied() {
		new(TicketUpdateParser).Parse(result, tx)
	}
}
